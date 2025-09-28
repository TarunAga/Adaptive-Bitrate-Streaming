package manifest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Manifest represents an HLS manifest
type Manifest struct {
	URL      string
	Bitrates []int
	Segments []string
}

// Fetch retrieves and parses a manifest from the given URL
func Fetch(url string) (*Manifest, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	
	return parseManifest(url, string(body))
}

// parseManifest parses HLS manifest content
func parseManifest(url, content string) (*Manifest, error) {
	manifest := &Manifest{
		URL: url,
	}
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Parse bitrate information
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			// Extract bandwidth from the line
			if strings.Contains(line, "BANDWIDTH=") {
				// Simple parsing - in real implementation, use proper parser
				parts := strings.Split(line, "BANDWIDTH=")
				if len(parts) > 1 {
					bandwidthStr := strings.Split(parts[1], ",")[0]
					var bandwidth int
					fmt.Sscanf(bandwidthStr, "%d", &bandwidth)
					manifest.Bitrates = append(manifest.Bitrates, bandwidth)
				}
			}
		}
		
		// Parse segment URLs
		if !strings.HasPrefix(line, "#") && line != "" {
			manifest.Segments = append(manifest.Segments, line)
		}
	}
	
	return manifest, nil
}
