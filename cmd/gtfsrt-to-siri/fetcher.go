package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// fetcher handles fetching GTFS-RT data from URLs or local files.
// This is CLI-specific logic and is not part of the core library.
type fetcher struct {
	httpClient *http.Client
}

// newFetcher creates a new fetcher for GTFS-RT data
func newFetcher() *fetcher {
	return &fetcher{
		httpClient: &http.Client{},
	}
}

// fetch fetches a single GTFS-RT feed from a URL or file path and returns raw protobuf bytes.
// Supports both HTTP URLs and local file paths.
// Returns nil if urlOrPath is empty (allows optional feeds).
func (f *fetcher) fetch(urlOrPath string) ([]byte, error) {
	if urlOrPath == "" {
		return nil, nil
	}

	// Check if it's a local file path
	if !strings.HasPrefix(urlOrPath, "http://") && !strings.HasPrefix(urlOrPath, "https://") {
		return os.ReadFile(urlOrPath)
	}

	// HTTP fetch
	resp, err := f.httpClient.Get(urlOrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", urlOrPath, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, urlOrPath)
	}

	return io.ReadAll(resp.Body)
}

// fetchAll fetches all three GTFS-RT feeds (trip updates, vehicle positions, service alerts).
// Supports both HTTP URLs and local file paths.
// Empty paths are skipped and return nil for that feed (allows optional feeds).
func (f *fetcher) fetchAll(tripUpdatesPath, vehiclePositionsPath, serviceAlertsPath string) ([]byte, []byte, []byte, error) {
	tu, err := f.fetch(tripUpdatesPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("trip updates: %w", err)
	}

	vp, err := f.fetch(vehiclePositionsPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("vehicle positions: %w", err)
	}

	sa, err := f.fetch(serviceAlertsPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("service alerts: %w", err)
	}

	return tu, vp, sa, nil
}
