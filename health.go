package gtfsrtsiri

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Status                  string `json:"status"`
	LatestGTFSRealtimeEpoch int64  `json:"latest_gtfsrt_epoch"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := healthResponse{
		Status:                  "ok",
		LatestGTFSRealtimeEpoch: 0,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
