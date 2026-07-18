package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type HTTPClient struct {
	client *http.Client

	baseURL string

	// retry settings
	maxRetries int
	retryDelay time.Duration
}

func NewHTTPClient(baseURL string, timeout time.Duration, maxRetries int, retryDelay time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	if retryDelay == 0 {
		retryDelay = 100 * time.Millisecond
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:    strings.TrimRight(baseURL, "/"),
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (c *HTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.do(req)
}

func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(path, "/")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.do(req)
}

func (c *HTTPClient) do(req *http.Request) (*http.Response, error) {
	var lastErr error

	delay := c.retryDelay

	for attempt := 0; attempt <= c.maxRetries; attempt++ {

		if attempt > 0 {

			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()

			case <-time.After(delay):
			}

			delay *= 2
		}

		resp, err := c.client.Do(req)

		if err != nil {

			lastErr = err

			if !isRetryableError(err) {
				return nil, fmt.Errorf(
					"execute request: %w",
					err,
				)
			}

			continue
		}

		if err := checkResponse(resp); err != nil {

			resp.Body.Close()

			lastErr = err

			if !isRetryableStatus(resp.StatusCode) {
				return nil, err
			}

			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf(
		"max retries exceeded: %w",
		lastErr,
	)
}

func checkResponse(resp *http.Response) error {

	if resp.StatusCode >= http.StatusBadRequest {

		return fmt.Errorf(
			"remote service returned status %d: %s",
			resp.StatusCode,
			resp.Status,
		)
	}

	return nil
}

func isRetryableStatus(status int) bool {

	switch status {
	case
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:

		return true
	}

	return false
}

func isRetryableError(err error) bool {

	var netErr net.Error

	if errors.As(err, &netErr) {

		return netErr.Timeout() || netErr.Temporary()
	}

	return false
}
