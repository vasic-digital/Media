package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolution_String(t *testing.T) {
	tests := []struct {
		name     string
		res      *Resolution
		expected string
	}{
		{"4K", &Resolution{3840, 2160}, "4K/UHD"},
		{"1080p", &Resolution{1920, 1080}, "1080p"},
		{"720p", &Resolution{1280, 720}, "720p"},
		{"480p", &Resolution{720, 480}, "480p"},
		{"custom", &Resolution{640, 360}, "640x360"},
		{"nil", nil, "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.res.String())
		})
	}
}

func TestResolution_Common(t *testing.T) {
	r := &Resolution{1920, 1080}
	assert.Equal(t, "1080p", r.Common())
}

func TestQualityInfo_IsBetterThan(t *testing.T) {
	high := &QualityInfo{QualityScore: 100}
	low := &QualityInfo{QualityScore: 50}
	var nilQuality *QualityInfo

	assert.True(t, high.IsBetterThan(low))
	assert.False(t, low.IsBetterThan(high))
	assert.False(t, low.IsBetterThan(low))
	assert.True(t, high.IsBetterThan(nilQuality))
	assert.False(t, nilQuality.IsBetterThan(high))
}

func TestQualityInfo_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		qi       QualityInfo
		expected string
	}{
		{
			"with profile",
			QualityInfo{QualityProfile: "BluRay 4K"},
			"BluRay 4K",
		},
		{
			"with resolution",
			QualityInfo{Resolution: &Resolution{1920, 1080}},
			"1080p",
		},
		{
			"unknown",
			QualityInfo{},
			"Unknown",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.qi.DisplayName())
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{1099511627776, "1.00 TB"},
		{5368709120, "5.00 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, FormatFileSize(tc.bytes))
		})
	}
}

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"The Matrix", "matrix"},
		{"A Beautiful Mind", "beautiful mind"},
		{"An Inconvenient Truth", "inconvenient truth"},
		{"Inception", "inception"},
		{"The.Dark.Knight", "dark knight"},
		{"Star_Wars_A_New_Hope", "star wars a new hope"},
		{"  Extra  Spaces  ", "extra spaces"},
		{"Mixed-Separators.And_Underscores", "mixed separators and underscores"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, NormalizeTitle(tc.input))
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_file", "normal_file"},
		{"file:with:colons", "file_with_colons"},
		{"file<with>brackets", "file_with_brackets"},
		{"file\"with\"quotes", "file_with_quotes"},
		{"file/with\\slashes", "file_with_slashes"},
		{"file|with|pipes", "file_with_pipes"},
		{"file?with*wildcards", "file_with_wildcards"},
		{".leading_dot", "leading_dot"},
		{"trailing_dot.", "trailing_dot"},
		{"  spaces  ", "spaces"},
		{"", "unnamed"},
		{"...", "unnamed"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, SanitizeFilename(tc.input))
		})
	}
}

func TestFileInfo(t *testing.T) {
	fi := FileInfo{
		Name:      "movie.mkv",
		Path:      "/data/movies/movie.mkv",
		Size:      1073741824,
		Extension: ".mkv",
		IsDir:     false,
	}

	assert.Equal(t, "movie.mkv", fi.Name)
	assert.Equal(t, "/data/movies/movie.mkv", fi.Path)
	assert.Equal(t, int64(1073741824), fi.Size)
	assert.Equal(t, ".mkv", fi.Extension)
	assert.False(t, fi.IsDir)
}
