// Package metadata provides common types and helper functions for
// working with media metadata across the detection, analysis, and
// provider packages.
package metadata

import (
	"fmt"
	"strings"
)

// QualityInfo represents quality information about a media file.
type QualityInfo struct {
	Resolution     *Resolution
	Bitrate        int
	VideoCodec     string
	AudioCodec     string
	FrameRate      float64
	HDR            bool
	Source         string // BluRay, WEB-DL, DVDRip, etc.
	QualityProfile string
	QualityScore   int
}

// Resolution represents video resolution dimensions.
type Resolution struct {
	Width  int
	Height int
}

// String returns a human-readable resolution string.
func (r *Resolution) String() string {
	if r == nil {
		return "unknown"
	}
	switch {
	case r.Height >= 2160:
		return "4K/UHD"
	case r.Height >= 1080:
		return "1080p"
	case r.Height >= 720:
		return "720p"
	case r.Height >= 480:
		return "480p"
	default:
		return fmt.Sprintf("%dx%d", r.Width, r.Height)
	}
}

// Common returns a common display name for the resolution.
func (r *Resolution) Common() string {
	return r.String()
}

// IsBetterThan compares two QualityInfo values and returns true if
// the receiver has a higher quality score than other.
func (qi *QualityInfo) IsBetterThan(other *QualityInfo) bool {
	if qi == nil || other == nil {
		return qi != nil
	}
	return qi.QualityScore > other.QualityScore
}

// DisplayName returns a human-readable quality name.
func (qi *QualityInfo) DisplayName() string {
	if qi.QualityProfile != "" {
		return qi.QualityProfile
	}
	if qi.Resolution != nil {
		return qi.Resolution.String()
	}
	return "Unknown"
}

// FileInfo represents basic file information for analysis.
type FileInfo struct {
	Name      string
	Path      string
	Size      int64
	Extension string
	IsDir     bool
}

// FormatFileSize formats a file size in bytes to a human-readable string.
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// NormalizeTitle cleans and normalizes a media title for comparison.
func NormalizeTitle(title string) string {
	// Convert to lowercase
	normalized := strings.ToLower(title)

	// Replace common separators with spaces
	replacer := strings.NewReplacer(
		".", " ",
		"_", " ",
		"-", " ",
	)
	normalized = replacer.Replace(normalized)

	// Remove leading/trailing articles for sorting
	articles := []string{"the ", "a ", "an "}
	for _, article := range articles {
		if strings.HasPrefix(normalized, article) {
			normalized = strings.TrimPrefix(normalized, article)
			break
		}
	}

	// Collapse whitespace
	fields := strings.Fields(normalized)
	normalized = strings.Join(fields, " ")

	return strings.TrimSpace(normalized)
}

// SanitizeFilename removes or replaces characters that are invalid
// in filenames across common operating systems.
func SanitizeFilename(name string) string {
	// Characters not allowed in Windows filenames
	invalid := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Remove leading/trailing whitespace and dots
	result = strings.Trim(result, " .")

	if result == "" {
		return "unnamed"
	}

	return result
}
