package steam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// BaseURL in Steam requset
	DefaultBaseURL = "https://api.steampowered.com"
	// Maximum time allowed for single request
	DefaultTimeout = 10 * time.Second
)

// Client calls the Steam Web API
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type Option func(*Client)

// overrides the API base URL
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("steam: API key is required")
	}
	c := &Client{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Steam error :
//   - bad/expired key         → HTTP 401, body may carry response.error
//   - private/unknown profile → HTTP 200 with an empty players slice (NOT an error)
type playerSummariesResponse struct {
	Response struct {
		Players []PlayerSummary `json:"players"`
		Error   *steamError     `json:"error,omitempty"`
	} `json:"response"`
}

type steamError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// fetches profile summaries for the given 64-bit Steam IDs
// Private or unknown IDs are simply absent from the result (this is normal,
// not an error).
// An error means the call itself failed (bad key, network, etc.)
func (c *Client) GetPlayerSummaries(ctx context.Context, steamIDs []string) ([]PlayerSummary, error) {
	if len(steamIDs) == 0 {
		return nil, nil
	}

	requestURL := fmt.Sprintf("%s/ISteamUser/GetPlayerSummaries/v0002/", c.baseURL)
	q := url.Values{}
	q.Set("key", c.apiKey)
	q.Set("steamids", strings.Join(steamIDs, ","))

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.URL.RawQuery = q.Encode()

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam: GetPlayerSummaries HTTP %d: %s",
			response.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed playerSummariesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("steam: decode GetPlayerSummaries: %w", err)
	}
	if parsed.Response.Error != nil {
		return nil, fmt.Errorf("steam: API error %d: %s",
			parsed.Response.Error.Code, parsed.Response.Error.Message)
	}
	return parsed.Response.Players, nil
}
