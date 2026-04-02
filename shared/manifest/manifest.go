package manifest

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// Base contains the common fields shared by all platform manifests.
type Base struct {
	Channel         string `json:"channel"`
	RocqVersion     string `json:"rocq_version"`
	PlatformRelease string `json:"platform_release"`
}

// Load reads a manifest file from an embedded filesystem, unmarshals it
// using the provided parse function, and returns the result.
func Load[T any](fsys fs.FS, path string, parse func([]byte) (*T, error)) (*T, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	return parse(data)
}

// UnmarshalBase extracts the common base fields from raw JSON manifest bytes.
func UnmarshalBase(data []byte) (*Base, error) {
	var b Base
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parse manifest base: %w", err)
	}
	return &b, nil
}
