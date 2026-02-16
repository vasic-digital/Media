// Package analyzer provides media file metadata extraction and analysis.
package analyzer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"digital.vasic.media/pkg/detector"
)

// Metadata contains extracted media metadata.
type Metadata struct {
	Title      string
	Year       int
	Genre      []string
	Duration   int // seconds
	Resolution string
	Codec      string
	Bitrate    int
	FileSize   int64
	MediaType  detector.MediaType
	Tags       map[string]string
}

// Analyzer extracts metadata from media files.
type Analyzer interface {
	Analyze(path string) (*Metadata, error)
	SupportedTypes() []detector.MediaType
}

// FilenameAnalyzer extracts metadata from filenames using patterns.
type FilenameAnalyzer struct {
	engine *detector.Engine
}

// NewFilenameAnalyzer creates a new FilenameAnalyzer with a detection engine.
func NewFilenameAnalyzer() *FilenameAnalyzer {
	return &FilenameAnalyzer{
		engine: detector.NewEngine(),
	}
}

// Analyze extracts metadata from the given file path based on filename patterns.
func (a *FilenameAnalyzer) Analyze(path string) (*Metadata, error) {
	filename := filepath.Base(path)
	if filename == "" || filename == "." {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	det := a.engine.Detect(filename)
	if det == nil {
		return nil, fmt.Errorf("detection failed for: %s", filename)
	}

	meta := &Metadata{
		Title:     det.Name,
		Year:      det.Year,
		MediaType: det.Type,
		Tags:      make(map[string]string),
	}

	// Copy tags from detection
	for k, v := range det.Tags {
		meta.Tags[k] = v
	}

	// Extract resolution from tags or filename
	if res, ok := det.Tags["resolution"]; ok {
		meta.Resolution = res
	} else {
		meta.Resolution = extractResolution(filename)
	}

	// Extract codec from tags or filename
	if codec, ok := det.Tags["video_codec"]; ok {
		meta.Codec = codec
	} else {
		meta.Codec = extractCodec(filename)
	}

	// For TV shows, add season/episode to tags
	if det.Type == detector.TypeTVShow {
		if det.Season > 0 {
			meta.Tags["season"] = strconv.Itoa(det.Season)
		}
		if det.Episode > 0 {
			meta.Tags["episode"] = strconv.Itoa(det.Episode)
		}
	}

	// For music, extract artist
	if det.Type == detector.TypeMusic {
		if artist, ok := det.Tags["artist"]; ok {
			meta.Tags["artist"] = artist
		}
	}

	// For books, extract author
	if det.Type == detector.TypeBook {
		if author, ok := det.Tags["author"]; ok {
			meta.Tags["author"] = author
		}
	}

	return meta, nil
}

// SupportedTypes returns the media types this analyzer can handle.
func (a *FilenameAnalyzer) SupportedTypes() []detector.MediaType {
	return []detector.MediaType{
		detector.TypeMovie,
		detector.TypeTVShow,
		detector.TypeMusic,
		detector.TypeBook,
		detector.TypePhoto,
		detector.TypeSoftware,
	}
}

// extractResolution attempts to find resolution info in a filename.
func extractResolution(filename string) string {
	lower := strings.ToLower(filename)

	patterns := []struct {
		needle string
		label  string
	}{
		{"2160p", "2160p"},
		{"4k", "2160p"},
		{"uhd", "2160p"},
		{"1080p", "1080p"},
		{"720p", "720p"},
		{"480p", "480p"},
	}

	for _, p := range patterns {
		if strings.Contains(lower, p.needle) {
			return p.label
		}
	}

	return ""
}

// extractCodec attempts to find codec info in a filename.
func extractCodec(filename string) string {
	lower := strings.ToLower(filename)

	codecPatterns := []struct {
		pattern *regexp.Regexp
		label   string
	}{
		{regexp.MustCompile(`(?i)\b(x265|h\.?265|hevc)\b`), "H.265"},
		{regexp.MustCompile(`(?i)\b(x264|h\.?264|avc)\b`), "H.264"},
		{regexp.MustCompile(`(?i)\bav1\b`), "AV1"},
		{regexp.MustCompile(`(?i)\bvp9\b`), "VP9"},
	}

	for _, cp := range codecPatterns {
		if cp.pattern.MatchString(lower) {
			return cp.label
		}
	}

	return ""
}
