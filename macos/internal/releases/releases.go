package releases

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/justme0606/rocq-bootstrap/macos/internal/manifest"
)

const (
	releasesURL = "https://api.github.com/repos/rocq-prover/platform/releases"
	releaseURL  = "https://api.github.com/repos/rocq-prover/platform/releases/tags/"
)

type ghRelease struct {
	TagName string `json:"tag_name"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ghReleaseDetail struct {
	TagName string    `json:"tag_name"`
	Body    string    `json:"body"`
	Assets  []ghAsset `json:"assets"`
}

// FetchReleases returns available release tags from GitHub, filtered to exclude
// old "v" prefixed tags.
func FetchReleases() ([]string, error) {
	resp, err := http.Get(releasesURL + "?per_page=30")
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch releases: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read releases body: %w", err)
	}

	var releases []ghRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("parse releases: %w", err)
	}

	var tags []string
	for _, r := range releases {
		if !strings.HasPrefix(r.TagName, "v") {
			tags = append(tags, r.TagName)
		}
	}

	return tags, nil
}

var versionRe = regexp.MustCompile(`\*\*(?:Rocq|Coq)\s+(\d+\.\d+\.\d+)\*\*`)

func inferRocqVersion(body string) string {
	if m := versionRe.FindStringSubmatch(body); m != nil {
		return m[1]
	}
	return ""
}

func findSignedDMG(assets []ghAsset) (string, string) {
	for _, a := range assets {
		if strings.HasPrefix(a.Name, "signed_") && strings.HasSuffix(a.Name, ".dmg") {
			// Prefer non-intel DMG (arm64)
			if !strings.Contains(strings.ToLower(a.Name), "intel") &&
				!strings.Contains(strings.ToLower(a.Name), "x86_64") &&
				!strings.Contains(strings.ToLower(a.Name), "amd64") {
				return a.BrowserDownloadURL, a.Name
			}
		}
	}
	// Fallback: any signed DMG
	for _, a := range assets {
		if strings.HasPrefix(a.Name, "signed_") && strings.HasSuffix(a.Name, ".dmg") {
			return a.BrowserDownloadURL, a.Name
		}
	}
	return "", ""
}

// FetchManifestForTag fetches a specific release from GitHub and builds a macOS manifest.
func FetchManifestForTag(tag string) (*manifest.Manifest, error) {
	resp, err := http.Get(releaseURL + tag)
	if err != nil {
		return nil, fmt.Errorf("fetch release %s: %w", tag, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch release %s: HTTP %d", tag, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read release body: %w", err)
	}

	var rel ghReleaseDetail
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("parse release: %w", err)
	}

	rocqVersion := inferRocqVersion(rel.Body)
	if rocqVersion == "" {
		return nil, fmt.Errorf("could not infer Rocq version from release %s body", tag)
	}

	dmgURL, _ := findSignedDMG(rel.Assets)
	if dmgURL == "" {
		return nil, fmt.Errorf("no signed .dmg asset found for release %s", tag)
	}

	m := &manifest.Manifest{
		Channel:         "stable",
		RocqVersion:     rocqVersion,
		PlatformRelease: tag,
		Assets: manifest.Assets{
			MacOS: struct {
				ARM64 manifest.Asset `json:"arm64"`
			}{
				ARM64: manifest.Asset{
					Type: "dmg",
					URL:  dmgURL,
				},
			},
		},
	}

	return m, nil
}
