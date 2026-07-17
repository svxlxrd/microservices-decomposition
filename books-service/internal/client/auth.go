package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AuthClient struct {
	httpClient *HTTPClient
	serviceKey string
	timeout    time.Duration
}

func NewAuthClient(baseURL string, timeout time.Duration, serviceKey string, requestTimeout time.Duration) *AuthClient {
	return &AuthClient{
		httpClient: NewHTTPClient(baseURL, timeout, 3, 100*time.Millisecond),
		serviceKey: serviceKey,
		timeout: requestTimeout,
	}
}

func (c *AuthClient) VerifyToken(ctx context.Context, token string) (*VerifyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := VerifyRequest{
		Token: token,
	}

	resp, err := c.httpClient.Post(
		ctx,
		"/internal/v1/auth/verify",
		req,
		map[string]string{
			"X-Service-Key": c.serviceKey},
	)

	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}
	defer resp.Body.Close()

	var result VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode verify response: %w", err)
	}

	return &result, nil
}

func (c *AuthClient) GetUsersByIDs(ctx context.Context, ids []string) ([]UserPublic, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := GetUsersByIDsRequest{
		IDs: ids,
	}

	resp, err := c.httpClient.Post(
		ctx,
		"/internal/v1/users/batch",
		req,
		map[string]string{
			"X-Service-Key": c.serviceKey},
	)

	if err != nil {
		return nil, fmt.Errorf("get users by ids: %w", err)
	}
	defer resp.Body.Close()

	var result GetUsersByIDsResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode users response: %w", err)
	}

	return result.Users, nil
}
