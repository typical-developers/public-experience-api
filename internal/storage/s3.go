package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/typical-developers/public-experience-api/internal/config"
)

const defaultRegion = "us-east-1"

func NewS3(ctx context.Context, config config.S3) (*s3.Client, error) {
	region := config.Region
	if region == "" {
		region = defaultRegion
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AccessKey,
			config.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.Endpoint)
		o.UsePathStyle = true

		// Newer SDK versions add CRC32 checksums to every upload by default.
		// Several S3-compatible providers reject or mishandle these, so only
		// send or check them when the operation requires it.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return client, nil
}

func Ping(ctx context.Context, client *s3.Client, bucket string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		return fmt.Errorf("reach bucket %q: %w", bucket, err)
	}

	return nil
}
