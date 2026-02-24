package windows

import "embed"

// EmbeddedManifest contains the manifest/latest.json file.
//
//go:embed embedded/manifest/latest.json
var EmbeddedManifest embed.FS

// EmbeddedTemplates contains workspace template files.
//
//go:embed embedded/templates/*
var EmbeddedTemplates embed.FS

// EmbeddedIcon contains the application icon.
//
//go:embed embedded/icon/rocq-icon.png
var EmbeddedIcon []byte
