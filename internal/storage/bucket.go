package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/errgroup"
)

// ErrNotFound is returned by Get and Head when the object does not exist.
var ErrNotFound = errors.New("storage: object not found")

const (
	defaultConcurrency = 8
	// The most keys S3 allows in a single DeleteObjects request.
	maxDeleteBatch = 1000
)

// Object is a file to upload.
type Object struct {
	Key          string
	Body         []byte
	ContentType  string
	CacheControl string
	Metadata     map[string]string
}

// ObjectInfo describes a stored object.
type ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	CacheControl string
	LastModified time.Time
	Metadata     map[string]string
}

// BatchError reports which keys failed in a PutMany or DeleteMany call.
// Keys that are not in Errors succeeded.
type BatchError struct {
	Op     string
	Errors map[string]error
}

func (e *BatchError) Error() string {
	keys := make([]string, 0, len(e.Errors))
	for key := range e.Errors {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	preview := keys
	if len(preview) > 3 {
		preview = append(preview[:3:3], fmt.Sprintf("and %d more", len(keys)-3))
	}

	return fmt.Sprintf("storage: %s failed for %d object(s): %s",
		e.Op, len(e.Errors), strings.Join(preview, ", "))
}

// Failed returns the error for one key, or nil if that key succeeded.
func (e *BatchError) Failed(key string) error {
	return e.Errors[key]
}

// Bucket wraps an S3 client for a single bucket.
type Bucket struct {
	client      *s3.Client
	name        string
	concurrency int
}

type BucketOption func(*Bucket)

// WithConcurrency sets how many uploads PutMany runs at once.
func WithConcurrency(n int) BucketOption {
	return func(b *Bucket) {
		if n > 0 {
			b.concurrency = n
		}
	}
}

func NewBucket(client *s3.Client, name string, opts ...BucketOption) *Bucket {
	b := &Bucket{
		client:      client,
		name:        name,
		concurrency: defaultConcurrency,
	}
	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *Bucket) Name() string       { return b.name }
func (b *Bucket) Client() *s3.Client { return b.client }

// Ping checks that the bucket exists and the credentials can reach it.
func (b *Bucket) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := b.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(b.name),
	}); err != nil {
		return fmt.Errorf("reach bucket %q: %w", b.name, err)
	}

	return nil
}

// Put uploads a single object.
func (b *Bucket) Put(ctx context.Context, obj Object) error {
	if obj.Key == "" {
		return errors.New("storage: object key is empty")
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(b.name),
		Key:           aws.String(obj.Key),
		Body:          bytes.NewReader(obj.Body),
		ContentLength: aws.Int64(int64(len(obj.Body))),
		Metadata:      obj.Metadata,
	}
	if obj.ContentType != "" {
		input.ContentType = aws.String(obj.ContentType)
	}
	if obj.CacheControl != "" {
		input.CacheControl = aws.String(obj.CacheControl)
	}

	if _, err := b.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("put %q: %w", obj.Key, err)
	}

	return nil
}

// PutMany uploads objects concurrently, up to the bucket's concurrency limit.
// A failed upload does not cancel the others. If any fail, a *BatchError
// lists them, so the caller can still record the ones that succeeded.
// Keys should be unique within a call.
func (b *Bucket) PutMany(ctx context.Context, objects []Object) error {
	var (
		mu     sync.Mutex
		failed = map[string]error{}
	)

	// A plain Group (not WithContext) so one failure doesn't cancel the rest.
	var g errgroup.Group
	g.SetLimit(b.concurrency)

	for _, obj := range objects {
		g.Go(func() error {
			err := ctx.Err()
			if err == nil {
				err = b.Put(ctx, obj)
			}
			if err != nil {
				mu.Lock()
				failed[obj.Key] = err
				mu.Unlock()
			}

			return nil
		})
	}
	_ = g.Wait()

	if len(failed) > 0 {
		return &BatchError{Op: "put", Errors: failed}
	}

	return nil
}

// DeleteMany deletes keys in batches of up to 1000.
// Deleting a key that does not exist counts as a success.
func (b *Bucket) DeleteMany(ctx context.Context, keys []string) error {
	failed := map[string]error{}

	for chunk := range slices.Chunk(keys, maxDeleteBatch) {
		if err := ctx.Err(); err != nil {
			for _, key := range chunk {
				failed[key] = err
			}
			continue
		}

		ids := make([]types.ObjectIdentifier, len(chunk))
		for i, key := range chunk {
			ids[i] = types.ObjectIdentifier{Key: aws.String(key)}
		}

		out, err := b.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(b.name),
			Delete: &types.Delete{
				Objects: ids,
				Quiet:   aws.Bool(true), // only return the keys that failed
			},
		})
		if err != nil {
			for _, key := range chunk {
				failed[key] = fmt.Errorf("delete objects: %w", err)
			}
			continue
		}

		for _, e := range out.Errors {
			failed[aws.ToString(e.Key)] = fmt.Errorf("%s: %s",
				aws.ToString(e.Code), aws.ToString(e.Message))
		}
	}

	if len(failed) > 0 {
		return &BatchError{Op: "delete", Errors: failed}
	}

	return nil
}

// List iterates over every object under prefix, fetching pages as needed.
// When a page fails, the error is yielded once and iteration stops.
//
//	for obj, err := range bucket.List(ctx, "roblox/") { ... }
func (b *Bucket) List(ctx context.Context, prefix string) iter.Seq2[ObjectInfo, error] {
	return func(yield func(ObjectInfo, error) bool) {
		p := s3.NewListObjectsV2Paginator(b.client, &s3.ListObjectsV2Input{
			Bucket: aws.String(b.name),
			Prefix: aws.String(prefix),
		})

		for p.HasMorePages() {
			page, err := p.NextPage(ctx)
			if err != nil {
				yield(ObjectInfo{}, fmt.Errorf("list s3://%s/%s: %w", b.name, prefix, err))
				return
			}

			for _, obj := range page.Contents {
				info := ObjectInfo{
					Key:          aws.ToString(obj.Key),
					Size:         aws.ToInt64(obj.Size),
					ETag:         strings.Trim(aws.ToString(obj.ETag), `"`),
					LastModified: aws.ToTime(obj.LastModified),
				}
				if !yield(info, nil) {
					return
				}
			}
		}
	}
}

// Get returns the object's body and info. The caller must close the body.
// Returns ErrNotFound if the key does not exist.
func (b *Bucket) Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	out, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, ObjectInfo{}, ErrNotFound
		}
		return nil, ObjectInfo{}, fmt.Errorf("get %q: %w", key, err)
	}

	return out.Body, ObjectInfo{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		ETag:         strings.Trim(aws.ToString(out.ETag), `"`),
		ContentType:  aws.ToString(out.ContentType),
		CacheControl: aws.ToString(out.CacheControl),
		LastModified: aws.ToTime(out.LastModified),
		Metadata:     out.Metadata,
	}, nil
}

// Head returns the object's info without downloading it.
// Returns ErrNotFound if the key does not exist.
func (b *Bucket) Head(ctx context.Context, key string) (ObjectInfo, error) {
	out, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return ObjectInfo{}, ErrNotFound
		}
		return ObjectInfo{}, fmt.Errorf("head %q: %w", key, err)
	}

	return ObjectInfo{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		ETag:         strings.Trim(aws.ToString(out.ETag), `"`),
		ContentType:  aws.ToString(out.ContentType),
		CacheControl: aws.ToString(out.CacheControl),
		LastModified: aws.ToTime(out.LastModified),
		Metadata:     out.Metadata,
	}, nil
}

// isNotFound reports whether err means the object is missing.
// A missing bucket is a config error, so it is not treated as not found.
func isNotFound(err error) bool {
	var noSuchBucket *types.NoSuchBucket
	if errors.As(err, &noSuchBucket) {
		return false
	}

	var noSuchKey *types.NoSuchKey
	var notFound *types.NotFound
	if errors.As(err, &noSuchKey) || errors.As(err, &notFound) {
		return true
	}

	// Some S3-compatible providers return a bare 404 with no typed error.
	var re *awshttp.ResponseError
	return errors.As(err, &re) && re.HTTPStatusCode() == http.StatusNotFound
}
