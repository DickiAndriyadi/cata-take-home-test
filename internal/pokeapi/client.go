package pokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cata-take-home-test/internal/util"

	"go.uber.org/zap"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
	maxRetries int
	backoff    util.BackoffStrategy
}

func NewClient(baseURL string, timeout time.Duration, maxRetries int, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger:     logger,
		maxRetries: maxRetries,
		backoff:    util.NewExponentialBackoff(500*time.Millisecond, 5*time.Second),
	}
}

type PokemonListResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []PokemonResult `json:"results"`
}

type PokemonResult struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokemonDetail struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}

// FetchPokemonList fetches list of pokemon with given limit
func (c *Client) FetchPokemonList(ctx context.Context, limit int) ([]PokemonResult, error) {
	url := fmt.Sprintf("%s/pokemon?limit=%d", c.baseURL, limit)
	var result PokemonListResponse
	err := c.getWithRetry(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return result.Results, nil
}

// FetchPokemonDetails fetch pokemon details by URL
func (c *Client) FetchPokemonDetails(ctx context.Context, url string) (*PokemonDetail, error) {
	var detail PokemonDetail
	err := c.getWithRetry(ctx, url, &detail)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// getWithRetry performs HTTP GET with retry and backoff
func (c *Client) getWithRetry(ctx context.Context, url string, v interface{}) error {
	var lastErr error
	for i := 0; i < c.maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			c.logger.Error("Failed to create request", zap.Error(err))
			return err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logger.Warn("HTTP request failed, retrying", zap.Int("attempt", i+1), zap.Error(err))
			time.Sleep(c.backoff.Duration(i))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			// Server error, retry
			lastErr = fmt.Errorf("server error: %s", resp.Status)
			c.logger.Warn("Server error, retrying", zap.Int("attempt", i+1), zap.String("status", resp.Status))
			time.Sleep(c.backoff.Duration(i))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(bodyBytes, v); err != nil {
			c.logger.Error("Failed to unmarshal response", zap.Error(err))
			return err
		}
		return nil
	}
	return lastErr
}
