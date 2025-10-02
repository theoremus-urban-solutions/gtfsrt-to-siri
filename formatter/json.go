package formatter

import (
	"encoding/json"

	"mta/gtfsrt-to-siri/siri"
)

type responseBuilder struct{}

func newResponseBuilder() *responseBuilder { return &responseBuilder{} }

// NewResponseBuilder creates a new response builder for formatting SIRI responses
func NewResponseBuilder() *responseBuilder {
	return newResponseBuilder()
}

// BuildJSON serializes a SIRI response to JSON
func (rb *responseBuilder) BuildJSON(res *siri.SiriResponse) []byte {
	b, _ := json.Marshal(res)
	return b
}
