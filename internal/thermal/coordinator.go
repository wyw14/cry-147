package thermal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL *url.URL
	http    *http.Client
}

func NewClient(rawURL string, client *http.Client) (*Client, error) {
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &Client{baseURL: baseURL, http: client}, nil
}

func (client *Client) Reading(ctx context.Context, zone string) (Reading, error) {
	endpoint := client.baseURL.ResolveReference(&url.URL{Path: "/zones/" + url.PathEscape(zone)})
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return Reading{}, err
	}
	response, err := client.http.Do(request)
	if err != nil {
		return Reading{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Reading{}, fmt.Errorf("thermal zone %s returned %s", zone, response.Status)
	}
	defer response.Body.Close()
	var reading Reading
	if err := json.NewDecoder(response.Body).Decode(&reading); err != nil {
		return Reading{}, fmt.Errorf("decode thermal reading: %w", err)
	}
	if reading.Zone == "" {
		reading.Zone = zone
	}
	if reading.RecordedAt.IsZero() {
		reading.RecordedAt = time.Now().UTC()
	}
	return reading, nil
}
