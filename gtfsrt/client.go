package gtfsrt

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	gtfsrtpb "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

func fetchFeed(url string) (*gtfsrtpb.FeedMessage, error) {
	// basic retry with backoff and context timeout
	var lastErr error
	timeout := 5 * time.Second
	attempts := 3
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		// optional headers could be injected here (API keys, etc.)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			cancel()
			time.Sleep(time.Duration(i+1) * 250 * time.Millisecond)
			continue
		}
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 250 * time.Millisecond)
			continue
		}
		var fm gtfsrtpb.FeedMessage
		if err := proto.Unmarshal(b, &fm); err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 250 * time.Millisecond)
			continue
		}
		return &fm, nil
	}
	if lastErr == nil {
		lastErr = errors.New("failed to fetch feed")
	}
	return nil, lastErr
}

// translatedStringToText returns the best-effort text from a TranslatedString
func translatedStringToText(ts *gtfsrtpb.TranslatedString) string {
	if ts == nil || len(ts.Translation) == 0 {
		return ""
	}
	// Prefer entries with no language tag or first entry
	var first string
	for _, tr := range ts.Translation {
		if tr.Text != nil {
			if tr.Language == nil || *tr.Language == "" {
				return *tr.Text
			}
			if first == "" {
				first = *tr.Text
			}
		}
	}
	return first
}
