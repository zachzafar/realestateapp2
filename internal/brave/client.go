package brave

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

type SearchResult struct {
	Title       string
	URL         string
	Description string
}

type searchResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

func New(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey: apiKey,
	}
}

func (c *Client) Search(ctx context.Context, query string) ([]SearchResult, error) {
	endpoint := "https://api.search.brave.com/res/v1/web/search"

	params := url.Values{}
	params.Set("q", query)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Subscription-Token", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("brave search failed: %s", resp.Status)
	}

	var data searchResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(data.Web.Results))

	for _, item := range data.Web.Results {
		results = append(results, SearchResult{
			Title:       item.Title,
			URL:         item.URL,
			Description: item.Description,
		})
	}

	return results, nil

}
