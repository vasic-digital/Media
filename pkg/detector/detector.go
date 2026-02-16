// Package detector provides media file type detection based on
// file extensions, MIME types, and content analysis.
package detector

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MediaType represents a detected media type.
type MediaType string

const (
	TypeMovie    MediaType = "movie"
	TypeTVShow   MediaType = "tv_show"
	TypeMusic    MediaType = "music"
	TypeBook     MediaType = "book"
	TypePhoto    MediaType = "photo"
	TypeGame     MediaType = "game"
	TypeSoftware MediaType = "software"
	TypeUnknown  MediaType = "unknown"
)

// Detection represents a media detection result.
type Detection struct {
	Type       MediaType
	Confidence float64 // 0.0 to 1.0
	Name       string
	Year       int
	Season     int
	Episode    int
	Extension  string
	Tags       map[string]string
}

// Engine detects media types from filenames and paths.
type Engine struct {
	rules []Rule
}

// Rule defines a detection rule.
type Rule struct {
	Name     string
	Type     MediaType
	Match    func(filename string) bool
	Extract  func(filename string) *Detection
	Priority int
}

// NewEngine creates a new detection engine with default rules.
func NewEngine() *Engine {
	e := &Engine{
		rules: make([]Rule, 0),
	}
	e.registerDefaultRules()
	return e
}

// Detect analyzes a filename and returns detection results.
func (e *Engine) Detect(filename string) *Detection {
	// Sort rules by priority descending (higher priority first)
	sorted := make([]Rule, len(e.rules))
	copy(sorted, e.rules)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})

	var best *Detection
	var bestConfidence float64

	for _, rule := range sorted {
		if rule.Match(filename) {
			det := rule.Extract(filename)
			if det != nil && det.Confidence > bestConfidence {
				best = det
				bestConfidence = det.Confidence
			}
		}
	}

	if best == nil {
		return &Detection{
			Type:       TypeUnknown,
			Confidence: 0.0,
			Name:       filenameWithoutExt(filename),
			Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
			Tags:       make(map[string]string),
		}
	}

	return best
}

// AddRule adds a custom detection rule.
func (e *Engine) AddRule(rule Rule) {
	e.rules = append(e.rules, rule)
}

// registerDefaultRules sets up all built-in detection rules.
func (e *Engine) registerDefaultRules() {
	// TV Show rule has highest priority because it has specific patterns
	// (S01E02) that override generic video extension matching.
	e.AddRule(Rule{
		Name:     "tv_show_pattern",
		Type:     TypeTVShow,
		Priority: 100,
		Match:    matchTVShow,
		Extract:  extractTVShow,
	})

	// Video files (movie detection)
	e.AddRule(Rule{
		Name:     "video_extensions",
		Type:     TypeMovie,
		Priority: 50,
		Match:    matchVideoExtensions,
		Extract:  extractMovie,
	})

	// Audio files (music detection)
	e.AddRule(Rule{
		Name:     "audio_extensions",
		Type:     TypeMusic,
		Priority: 50,
		Match:    matchAudioExtensions,
		Extract:  extractMusic,
	})

	// Image files (photo detection)
	e.AddRule(Rule{
		Name:     "image_extensions",
		Type:     TypePhoto,
		Priority: 50,
		Match:    matchImageExtensions,
		Extract:  extractPhoto,
	})

	// Book files
	e.AddRule(Rule{
		Name:     "book_extensions",
		Type:     TypeBook,
		Priority: 50,
		Match:    matchBookExtensions,
		Extract:  extractBook,
	})

	// Software files
	e.AddRule(Rule{
		Name:     "software_extensions",
		Type:     TypeSoftware,
		Priority: 50,
		Match:    matchSoftwareExtensions,
		Extract:  extractSoftware,
	})
}

// Extension sets for each media type.
var (
	videoExtensions    = newExtSet(".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm")
	audioExtensions    = newExtSet(".mp3", ".flac", ".wav", ".aac", ".ogg", ".wma", ".m4a")
	imageExtensions    = newExtSet(".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".tiff")
	bookExtensions     = newExtSet(".pdf", ".epub", ".mobi", ".azw3", ".cbr", ".cbz")
	softwareExtensions = newExtSet(".exe", ".msi", ".dmg", ".deb", ".rpm", ".appimage")
)

// TV show patterns: S01E02, s01e02, 1x02, Season 1 Episode 2, etc.
var tvShowPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)S(\d{1,4})E(\d{1,4})`),
	regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,3})`),
	regexp.MustCompile(`(?i)Season\s*(\d{1,4})\s*Episode\s*(\d{1,4})`),
}

// yearPattern captures a 4-digit year in parentheses, brackets, or standalone.
var yearPattern = regexp.MustCompile(`[\(\[.]?(\d{4})[\)\].]?`)

// cleanupPattern removes common release info from titles.
var cleanupPattern = regexp.MustCompile(`(?i)\b(bluray|brrip|dvdrip|webrip|web-dl|webdl|hdtv|hdcam|720p|1080p|2160p|4k|uhd|x264|x265|h264|h265|hevc|avc|aac|dts|ac3|remux|proper|repack|internal|complete|season|series)\b`)

// Match functions.

func matchTVShow(filename string) bool {
	for _, pat := range tvShowPatterns {
		if pat.MatchString(filename) {
			ext := strings.ToLower(filepath.Ext(filename))
			if _, ok := videoExtensions[ext]; ok {
				return true
			}
			// Also match if no extension (directory name)
			if ext == "" {
				return true
			}
		}
	}
	return false
}

func matchVideoExtensions(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := videoExtensions[ext]
	return ok
}

func matchAudioExtensions(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := audioExtensions[ext]
	return ok
}

func matchImageExtensions(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := imageExtensions[ext]
	return ok
}

func matchBookExtensions(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := bookExtensions[ext]
	return ok
}

func matchSoftwareExtensions(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := softwareExtensions[ext]
	return ok
}

// Extract functions.

func extractTVShow(filename string) *Detection {
	det := &Detection{
		Type:       TypeTVShow,
		Confidence: 0.9,
		Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
		Tags:       make(map[string]string),
	}

	base := filenameWithoutExt(filename)

	// Try each TV show pattern
	for _, pat := range tvShowPatterns {
		matches := pat.FindStringSubmatchIndex(base)
		if matches == nil {
			continue
		}

		// Extract season and episode numbers
		submatch := pat.FindStringSubmatch(base)
		if len(submatch) >= 3 {
			det.Season, _ = strconv.Atoi(submatch[1])
			det.Episode, _ = strconv.Atoi(submatch[2])
			det.Tags["season"] = submatch[1]
			det.Tags["episode"] = submatch[2]
		}

		// Title is everything before the pattern match
		titlePart := base[:matches[0]]
		det.Name = cleanTitle(titlePart)
		det.Year = extractYear(titlePart)

		if det.Season > 0 && det.Episode > 0 {
			det.Confidence = 0.95
		}

		return det
	}

	return det
}

func extractMovie(filename string) *Detection {
	base := filenameWithoutExt(filename)
	det := &Detection{
		Type:       TypeMovie,
		Confidence: 0.7,
		Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
		Tags:       make(map[string]string),
	}

	det.Name = cleanTitle(base)
	det.Year = extractYear(base)

	if det.Year > 0 {
		det.Confidence = 0.8
	}

	// Extract quality hints into tags
	extractQualityTags(base, det.Tags)

	return det
}

func extractMusic(filename string) *Detection {
	base := filenameWithoutExt(filename)
	det := &Detection{
		Type:       TypeMusic,
		Confidence: 0.85,
		Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
		Tags:       make(map[string]string),
	}

	// Music files often have "Artist - Title" format
	parts := strings.SplitN(base, " - ", 2)
	if len(parts) == 2 {
		det.Tags["artist"] = strings.TrimSpace(parts[0])
		det.Name = strings.TrimSpace(parts[1])
		det.Confidence = 0.9
	} else {
		det.Name = cleanTitle(base)
	}

	det.Year = extractYear(base)

	return det
}

func extractPhoto(filename string) *Detection {
	base := filenameWithoutExt(filename)
	return &Detection{
		Type:       TypePhoto,
		Confidence: 0.85,
		Name:       base,
		Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
		Tags:       make(map[string]string),
	}
}

func extractBook(filename string) *Detection {
	base := filenameWithoutExt(filename)
	det := &Detection{
		Type:       TypeBook,
		Confidence: 0.85,
		Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
		Tags:       make(map[string]string),
	}

	// Books often have "Author - Title" or "Title (Year)" format
	parts := strings.SplitN(base, " - ", 2)
	if len(parts) == 2 {
		det.Tags["author"] = strings.TrimSpace(parts[0])
		det.Name = strings.TrimSpace(parts[1])
		det.Confidence = 0.9
	} else {
		det.Name = cleanTitle(base)
	}

	det.Year = extractYear(base)

	return det
}

func extractSoftware(filename string) *Detection {
	base := filenameWithoutExt(filename)
	return &Detection{
		Type:       TypeSoftware,
		Confidence: 0.85,
		Name:       cleanTitle(base),
		Year:       extractYear(base),
		Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
		Tags:       make(map[string]string),
	}
}

// Helper functions.

func newExtSet(exts ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	return m
}

func filenameWithoutExt(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}

func cleanTitle(raw string) string {
	// Remove quality/release info
	title := cleanupPattern.ReplaceAllString(raw, "")

	// Remove year in parentheses/brackets for cleaner title
	title = regexp.MustCompile(`[\(\[]?\d{4}[\)\]]?`).ReplaceAllString(title, "")

	// Replace dots, underscores, hyphens with spaces
	title = regexp.MustCompile(`[._\-]+`).ReplaceAllString(title, " ")

	// Collapse multiple spaces
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	return strings.TrimSpace(title)
}

func extractYear(s string) int {
	matches := yearPattern.FindAllStringSubmatch(s, -1)
	currentYear := time.Now().Year()

	for _, match := range matches {
		if len(match) >= 2 {
			y, err := strconv.Atoi(match[1])
			if err == nil && y >= 1900 && y <= currentYear+5 {
				return y
			}
		}
	}
	return 0
}

func extractQualityTags(s string, tags map[string]string) {
	lower := strings.ToLower(s)

	// Resolution
	if strings.Contains(lower, "2160p") || strings.Contains(lower, "4k") || strings.Contains(lower, "uhd") {
		tags["resolution"] = "2160p"
	} else if strings.Contains(lower, "1080p") {
		tags["resolution"] = "1080p"
	} else if strings.Contains(lower, "720p") {
		tags["resolution"] = "720p"
	} else if strings.Contains(lower, "480p") {
		tags["resolution"] = "480p"
	}

	// Source
	if strings.Contains(lower, "bluray") || strings.Contains(lower, "brrip") {
		tags["source"] = "BluRay"
	} else if strings.Contains(lower, "web-dl") || strings.Contains(lower, "webdl") {
		tags["source"] = "WEB-DL"
	} else if strings.Contains(lower, "webrip") {
		tags["source"] = "WEBRip"
	} else if strings.Contains(lower, "hdtv") {
		tags["source"] = "HDTV"
	} else if strings.Contains(lower, "dvdrip") {
		tags["source"] = "DVDRip"
	}

	// Video codec
	if strings.Contains(lower, "x265") || strings.Contains(lower, "h265") || strings.Contains(lower, "hevc") {
		tags["video_codec"] = "H.265"
	} else if strings.Contains(lower, "x264") || strings.Contains(lower, "h264") || strings.Contains(lower, "avc") {
		tags["video_codec"] = "H.264"
	}

	// Audio codec
	if strings.Contains(lower, "dts") {
		tags["audio_codec"] = "DTS"
	} else if strings.Contains(lower, "aac") {
		tags["audio_codec"] = "AAC"
	} else if strings.Contains(lower, "ac3") {
		tags["audio_codec"] = "AC3"
	}

	// HDR
	if strings.Contains(lower, "hdr") || strings.Contains(lower, "dolby.vision") {
		tags["hdr"] = "true"
	}
}
