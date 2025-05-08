package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	DefaultTimeout    = 20 * time.Second
	DefaultRetries    = 3
	DefaultRetryDelay = 500 * time.Millisecond
)

type Opts struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
	Logger     *zap.Logger
	Headers    map[string]string
}

func NewDefaultOpts(logger *zap.Logger) *Opts {
	return &Opts{
		Timeout:    DefaultTimeout,
		RetryDelay: DefaultRetryDelay,
		MaxRetries: DefaultRetries,
		Logger:     logger,
		Headers:    make(map[string]string),
	}
}

type Client struct {
	httpClient *http.Client
	options    *Opts
}

func (c *Client) requestRaw(
	ctx context.Context,
	method string,
	url string,
	body any,
) ([]byte, error) {
	if c.options.Logger == nil {
		return nil, fmt.Errorf("logger is required in http client, use nop logger if neccessary")
	}

	var reqBody io.Reader

	if body != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.options.Logger.Error("failed to encode request body", zap.Error(err))
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
		reqBody = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		c.options.Logger.Error("failed to create request", zap.Error(err))
		return nil, fmt.Errorf("failed to create create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	var lastError error
	for attempt := range c.options.MaxRetries {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.options.RetryDelay):
				// wait in case its not the first attempt
			}
		}

		if attempt > 0 && c.options.Logger != nil {
			c.options.Logger.Info("Retrying request",
				zap.String("url", url),
				zap.String("method", method),
				zap.Int("attempt", attempt),
				zap.Error(lastError))
		}

		response, err := c.httpClient.Do(req)
		if err != nil {
			lastError = err

			select {
			case <-ctx.Done():
				c.options.Logger.Error("request timed out", zap.Error(err))
				return nil, ctx.Err()
			default:
				continue
			}
		}

		defer response.Body.Close()

		respBody, err := io.ReadAll(response.Body)
		if err != nil {
			c.options.Logger.Error("error reading response body", zap.Error(err))
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if response.StatusCode < 400 {
			c.options.Logger.Info("Got 200 response", zap.Int("response-code", response.StatusCode))
			return respBody, nil
		} else if response.StatusCode < 500 {
			c.options.Logger.Error("Got 400 response", zap.Int("response-code", response.StatusCode),
				zap.Any("request-body", body),
				zap.Any("response-body", string(respBody)))
			return nil, fmt.Errorf("Got 400 response")
		} else {
			c.options.Logger.Error("Got 500 response will retry.", zap.Int("response-code", response.StatusCode),
				zap.Any("body", string(respBody)))
		}
	}

	c.options.Logger.Error("Request failed after all retries",
		zap.Error(lastError),
		zap.String("url", url),
		zap.String("method", method),
		zap.Int("max_retries", c.options.MaxRetries))

	return nil, fmt.Errorf(
		"request failed after %d attempts: %w",
		c.options.MaxRetries+1,
		lastError,
	)
}

func request[T any](
	ctx context.Context,
	method string,
	url string,
	body any,
	options *Opts,
	ignoreTls bool,
) (T, error) {
	var client *Client
	if ignoreTls {
		client = &Client{
			httpClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			},
			options: options,
		}
	} else {
		client = &Client{
			httpClient: http.DefaultClient,
			options:    options,
		}
	}

	var result T

	responseBody, err := client.requestRaw(ctx, method, url, body)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(responseBody, &result); err != nil {
		client.options.Logger.Error("error unmarshaling response body", zap.Error(err))
		return result, err
	}

	return result, nil
}

func GetRequest[T any](
	ctx context.Context,
	logger *zap.Logger,
	url string,
	ignoreTls bool,
) (T, error) {
	return request[T](ctx, http.MethodGet, url, nil, NewDefaultOpts(logger), ignoreTls)
}

func PostRequest[T any](
	ctx context.Context,
	logger *zap.Logger,
	url string,
	body any,
	ignoreTls bool,
) (T, error) {
	return request[T](ctx, http.MethodPost, url, body, NewDefaultOpts(logger), ignoreTls)
}
