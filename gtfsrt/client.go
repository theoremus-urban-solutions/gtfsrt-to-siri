package gtfsrt

import (
	"fmt"
	"io"
	"net/http"
)

// Client is a simple HTTP client for fetching GTFS-RT protobuf data.
// This is a CLI helper - library users should fetch data themselves.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new GTFS-RT HTTP client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

// Fetch fetches a single GTFS-RT feed from a URL and returns raw protobuf bytes.
// Returns nil if url is empty (allows optional feeds).
func (c *Client) Fetch(url string) ([]byte, error) {
	if url == "" {
		return nil, nil
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

// FetchAll fetches all three GTFS-RT feeds (trip updates, vehicle positions, service alerts).
// Empty URLs are skipped and return nil for that feed (allows optional feeds).
func (c *Client) FetchAll(tripUpdatesURL, vehiclePositionsURL, serviceAlertsURL string) ([]byte, []byte, []byte, error) {
	tu, err := c.Fetch(tripUpdatesURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("trip updates: %w", err)
	}

	vp, err := c.Fetch(vehiclePositionsURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("vehicle positions: %w", err)
	}

	sa, err := c.Fetch(serviceAlertsURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("service alerts: %w", err)
	}

	return tu, vp, sa, nil
}
