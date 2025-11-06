package gtfs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
)

// SerializeIndex encodes a GTFSIndex to bytes using gob encoding.
// This is useful for disk-based caching to avoid re-parsing GTFS static data.
//
// Example:
//
//	index, _ := gtfs.NewGTFSIndexFromBytes(zipBytes, "AGENCY")
//	data, err := gtfs.SerializeIndex(index)
//	if err != nil {
//	    // handle error
//	}
//	// Save to disk
//	os.WriteFile("/path/to/cache/index.gob", data, 0644)
//
// Thread safety: Safe for concurrent use once the index is fully constructed.
func SerializeIndex(index *GTFSIndex) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(index); err != nil {
		return nil, fmt.Errorf("failed to encode GTFSIndex: %w", err)
	}
	return buf.Bytes(), nil
}

// DeserializeIndex decodes a GTFSIndex from bytes using gob encoding.
// Use this to load a previously serialized index from disk cache.
//
// Example:
//
//	data, _ := os.ReadFile("/path/to/cache/index.gob")
//	index, err := gtfs.DeserializeIndex(data)
//	if err != nil {
//	    // Cache is corrupted or invalid, fetch fresh data
//	    index, _ = gtfs.NewGTFSIndexFromBytes(freshZipBytes, "AGENCY")
//	}
//
// Thread safety: The returned index is safe for concurrent read access.
func DeserializeIndex(data []byte) (*GTFSIndex, error) {
	buf := bytes.NewReader(data)
	decoder := gob.NewDecoder(buf)
	var index GTFSIndex
	if err := decoder.Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to decode GTFSIndex: %w", err)
	}
	return &index, nil
}

// SerializeIndexToFile writes a GTFSIndex to a file using gob encoding.
// This is a convenience wrapper around SerializeIndex for direct file I/O.
//
// Example:
//
//	index, _ := gtfs.NewGTFSIndexFromBytes(zipBytes, "AGENCY")
//	if err := gtfs.SerializeIndexToFile(index, "/cache/gtfs-index.gob"); err != nil {
//	    // handle error
//	}
func SerializeIndexToFile(index *GTFSIndex, filepath string) error {
	data, err := SerializeIndex(index)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// DeserializeIndexFromFile reads a GTFSIndex from a file using gob encoding.
// This is a convenience wrapper around DeserializeIndex for direct file I/O.
//
// Example:
//
//	index, err := gtfs.DeserializeIndexFromFile("/cache/gtfs-index.gob")
//	if err != nil {
//	    // Cache miss or corrupted, fetch fresh data
//	    index, _ = gtfs.NewGTFSIndexFromBytes(freshZipBytes, "AGENCY")
//	}
func DeserializeIndexFromFile(filepath string) (*GTFSIndex, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}
	return DeserializeIndex(data)
}

// SerializeIndexToWriter writes a GTFSIndex to an io.Writer using gob encoding.
// This provides maximum flexibility for custom storage backends (S3, MinIO, etc.).
//
// Example:
//
//	var buf bytes.Buffer
//	index, _ := gtfs.NewGTFSIndexFromBytes(zipBytes, "AGENCY")
//	if err := gtfs.SerializeIndexToWriter(index, &buf); err != nil {
//	    // handle error
//	}
//	// Upload buf.Bytes() to S3, MinIO, etc.
func SerializeIndexToWriter(index *GTFSIndex, w io.Writer) error {
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(index); err != nil {
		return fmt.Errorf("failed to encode GTFSIndex: %w", err)
	}
	return nil
}

// DeserializeIndexFromReader reads a GTFSIndex from an io.Reader using gob encoding.
// This provides maximum flexibility for custom storage backends (S3, MinIO, etc.).
//
// Example:
//
//	// Download from S3, MinIO, etc.
//	reader := bytes.NewReader(downloadedData)
//	index, err := gtfs.DeserializeIndexFromReader(reader)
//	if err != nil {
//	    // handle error
//	}
func DeserializeIndexFromReader(r io.Reader) (*GTFSIndex, error) {
	decoder := gob.NewDecoder(r)
	var index GTFSIndex
	if err := decoder.Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to decode GTFSIndex: %w", err)
	}
	return &index, nil
}
