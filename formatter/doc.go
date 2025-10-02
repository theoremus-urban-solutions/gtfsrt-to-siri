// Package formatter provides response wrapping and serialization for SIRI responses.
//
// This package is organized into:
// - wrapper.go: Response wrapping logic (ServiceDelivery, filtering, utilities)
// - json.go: JSON serialization
// - xml.go: XML serialization with proper escaping
//
// All serialization is done manually for performance and precise control over output format.
package formatter
