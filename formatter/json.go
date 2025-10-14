package formatter

import (
	"encoding/json"

	"github.com/theoremus-urban-solutions/gtfsrt-to-siri/utils"
)

type responseBuilder struct{}

func newResponseBuilder() *responseBuilder { return &responseBuilder{} }

// NewResponseBuilder creates a new response builder for formatting SIRI responses
func NewResponseBuilder() *responseBuilder {
	return newResponseBuilder()
}

// BuildJSON serializes a SIRI response to JSON
func (rb *responseBuilder) BuildJSON(res *utils.SiriResponse) []byte {
	b, _ := json.Marshal(res)
	return b
}
