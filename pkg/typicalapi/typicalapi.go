package typicalapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type Response struct {
	*http.Response
}

// Ok will check if the Request made returned an OK status
// ranging between 200-299.
func (r *Response) Ok() bool {
	return r.StatusCode >= 200 && r.StatusCode <= 299
}

type service struct {
	client *Client
}

type Client struct {
	http    *http.Client
	baseURL url.URL
	common  service

	Oaklands *oaklands
}

type ClientOpt func(c *Client)

// WithBaseURL will let you change the base url that the
// client uses.
func WithBaseURL(u url.URL) ClientOpt {
	return func(c *Client) {
		c.baseURL = u
	}
}

// WithHTTPClient will let you use a custom http client.
func WithHTTPClient(client *http.Client) ClientOpt {
	return func(c *Client) {
		c.http = client
	}
}

func New(opts ...ClientOpt) *Client {
	c := &Client{
		http: http.DefaultClient,
		baseURL: url.URL{
			Scheme: "https",
			// NOTE:
			// 		This is the *planned* url. Clients at the moment should use
			// 		[WithBaseURL] to override with the current placeholder deployment,
			// 		`public-staging.luckfire.dev`.
			Host: "public-experience.typical.gg",
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c.init()
}

func (c *Client) init() *Client {
	c.common.client = c

	c.Oaklands = (*oaklands)(&c.common)

	return c
}

// NewGETRequest creates a new GET request.
//
// At the moment, the public api only really supports GET requests.
func (c *Client) NewGETRequest(uStr string) (*http.Request, error) {
	u, err := c.baseURL.Parse(uStr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) send(ctx context.Context, req *http.Request) (*Response, error) {
	resp, err := c.http.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	response := &Response{resp}
	return response, nil
}

// Call will call a compiled endpoint.
func (c *Client) Call(ctx context.Context, e string, body, out any) (*Response, error) {
	req, err := c.NewGETRequest(e)
	if err != nil {
		return nil, err
	}

	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !resp.Ok() {
		apiErr := &Error{Response: resp}
		_ = json.NewDecoder(resp.Body).Decode(apiErr)

		return resp, apiErr
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp, err
		}
	}

	return resp, nil
}
