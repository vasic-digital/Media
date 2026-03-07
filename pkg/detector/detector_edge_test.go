package detector

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unicode Filenames ---

func TestDetector_UnicodeFilenames(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		name     string
		filename string
		wantType MediaType
	}{
		{
			name:     "CJK_characters_video",
			filename: "千と千尋の神隠し.2001.1080p.mkv",
			wantType: TypeMovie,
		},
		{
			name:     "Cyrillic_characters_video",
			filename: "Брат.1997.720p.BluRay.mkv",
			wantType: TypeMovie,
		},
		{
			name:     "Arabic_characters_video",
			filename: "فيلم عربي.2020.mp4",
			wantType: TypeMovie,
		},
		{
			name:     "Korean_characters_audio",
			filename: "방탄소년단 - Dynamite.mp3",
			wantType: TypeMusic,
		},
		{
			name:     "Japanese_characters_book",
			filename: "村上春樹 - ノルウェイの森.epub",
			wantType: TypeBook,
		},
		{
			name:     "Chinese_TV_show",
			filename: "三体.S01E01.2024.mkv",
			wantType: TypeTVShow,
		},
		{
			name:     "Mixed_unicode_and_latin",
			filename: "Amélie.2001.1080p.mkv",
			wantType: TypeMovie,
		},
		{
			name:     "Emoji_in_filename",
			filename: "🎬movie🎬.mkv",
			wantType: TypeMovie,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, tc.wantType, det.Type, "wrong type for %s", tc.filename)
			assert.NotNil(t, det.Tags, "tags should not be nil")
		})
	}
}

// --- Very Long Filename ---

func TestDetector_VeryLongFilename(t *testing.T) {
	e := NewEngine()

	// Build a filename with 500+ characters
	longName := strings.Repeat("A", 500) + ".1999.1080p.BluRay.x264.mkv"
	det := e.Detect(longName)
	require.NotNil(t, det)
	assert.Equal(t, TypeMovie, det.Type)
	assert.Equal(t, 1999, det.Year)
	assert.True(t, det.Confidence >= 0.7)
	assert.NotEmpty(t, det.Name)

	// Very long filename without known extension
	longUnknown := strings.Repeat("B", 600) + ".xyz"
	det2 := e.Detect(longUnknown)
	require.NotNil(t, det2)
	assert.Equal(t, TypeUnknown, det2.Type)

	// Very long path with short filename
	longPath := "/" + strings.Repeat("dir/", 100) + "movie.mkv"
	det3 := e.Detect(longPath)
	require.NotNil(t, det3)
	assert.Equal(t, TypeMovie, det3.Type)
}

// --- Multiple Resolution Tags ---

func TestDetector_MultipleResolutionTags(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		name           string
		filename       string
		wantResolution string
	}{
		{
			name:           "2160p_and_4k",
			filename:       "Movie.2160p.4k.mkv",
			wantResolution: "2160p",
		},
		{
			name:           "4k_and_uhd",
			filename:       "Movie.4K.UHD.HDR.mkv",
			wantResolution: "2160p",
		},
		{
			name:           "1080p_and_720p_2160p_wins",
			filename:       "Movie.1080p.720p.mkv",
			wantResolution: "1080p",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, TypeMovie, det.Type)
			assert.Equal(t, tc.wantResolution, det.Tags["resolution"])
		})
	}
}

// --- Special Characters ---

func TestDetector_SpecialCharacters(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		name     string
		filename string
		wantType MediaType
	}{
		{
			name:     "brackets",
			filename: "[GROUP] Movie Name (2020) [1080p].mkv",
			wantType: TypeMovie,
		},
		{
			name:     "parentheses_in_name",
			filename: "Movie (Directors Cut) (2015).mp4",
			wantType: TypeMovie,
		},
		{
			name:     "dots_only_separator",
			filename: "The.Movie.Name.2019.BluRay.mkv",
			wantType: TypeMovie,
		},
		{
			name:     "dashes_separator",
			filename: "The-Movie-Name-2019-BluRay.mkv",
			wantType: TypeMovie,
		},
		{
			name:     "underscores_separator",
			filename: "The_Movie_Name_2019_BluRay.mkv",
			wantType: TypeMovie,
		},
		{
			name:     "mixed_separators",
			filename: "The.Movie-Name_2019.BluRay.mkv",
			wantType: TypeMovie,
		},
		{
			name:     "ampersand_in_name",
			filename: "Tom & Jerry.2021.mkv",
			wantType: TypeMovie,
		},
		{
			name:     "exclamation_in_name",
			filename: "Wow!.2020.mp4",
			wantType: TypeMovie,
		},
		{
			name:     "hash_in_name",
			filename: "Track #5.mp3",
			wantType: TypeMusic,
		},
		{
			name:     "plus_in_name",
			filename: "C++ Tutorial.pdf",
			wantType: TypeBook,
		},
		{
			name:     "at_sign_in_name",
			filename: "user@home.jpg",
			wantType: TypePhoto,
		},
		{
			name:     "curly_braces",
			filename: "Movie {2020}.mkv",
			wantType: TypeMovie,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, tc.wantType, det.Type, "unexpected type for %q", tc.filename)
			assert.NotNil(t, det.Tags)
		})
	}
}

// --- Empty Filename ---

func TestDetector_EmptyFilename(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		name     string
		filename string
	}{
		{"empty_string", ""},
		{"single_dot", "."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det, "detect should never return nil")
			assert.Equal(t, TypeUnknown, det.Type)
			assert.NotNil(t, det.Tags)
		})
	}
}

// --- Path Only, No Filename ---

func TestDetector_PathOnly_NoFilename(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		name     string
		path     string
		wantType MediaType
	}{
		{
			name:     "trailing_slash_returns_last_dir",
			path:     "/some/dir/",
			wantType: TypeUnknown, // filepath.Base("/some/dir/") = "dir"
		},
		{
			name:     "just_slash",
			path:     "/",
			wantType: TypeUnknown,
		},
		{
			name:     "directory_with_tv_pattern_no_video_ext",
			path:     "Show.S01E05",
			wantType: TypeUnknown, // S01E05 becomes the "extension" via filepath.Ext, no video ext match
		},
		{
			name:     "directory_without_extension",
			path:     "/movies/The Matrix (1999)/",
			wantType: TypeUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			det := e.Detect(tc.path)
			require.NotNil(t, det)
			assert.Equal(t, tc.wantType, det.Type)
		})
	}
}

// --- Year Extraction Edge Cases ---

func TestDetector_YearExtraction_EdgeCases(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name     string
		input    string
		wantYear int
	}{
		{
			name:     "year_1899_too_old",
			input:    "Movie.1899.mkv",
			wantYear: 0, // 1899 < 1900, should not be extracted
		},
		{
			name:     "year_1900_boundary",
			input:    "Movie.1900.mkv",
			wantYear: 1900, // 1900 is the minimum valid year
		},
		{
			name:     "year_far_future_invalid",
			input:    fmt.Sprintf("Movie.%d.mkv", currentYear+6),
			wantYear: 0, // Beyond currentYear+5
		},
		{
			name:     "year_near_future_valid",
			input:    fmt.Sprintf("Movie.%d.mkv", currentYear+4),
			wantYear: currentYear + 4, // Within currentYear+5
		},
		{
			name:     "number_1080_not_year",
			input:    "Movie.1080p.BluRay.mkv",
			wantYear: 0, // 1080 < 1900, not valid as year
		},
		{
			name:     "number_2160_not_year_in_2026",
			input:    "Movie.2160p.mkv",
			wantYear: 0, // 2160 > currentYear+5, not valid (as of 2026)
		},
		{
			name:     "year_in_parentheses",
			input:    "Movie (2020)",
			wantYear: 2020,
		},
		{
			name:     "year_in_brackets",
			input:    "Movie [2019]",
			wantYear: 2019,
		},
		{
			name:     "multiple_years_first_valid_wins",
			input:    "Movie.2001.Sequel.2010",
			wantYear: 2001,
		},
		{
			name:     "no_year_all_short_numbers",
			input:    "Movie.720.480",
			wantYear: 0,
		},
		{
			name:     "five_digit_number_not_year",
			input:    "Movie12001Name",
			wantYear: 0,
		},
		{
			name:     "current_year_valid",
			input:    fmt.Sprintf("Movie.%d.mkv", currentYear),
			wantYear: currentYear,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractYear(tc.input)
			assert.Equal(t, tc.wantYear, result, "input: %s", tc.input)
		})
	}
}

// TestDetector_2160AsYear documents behavior: 2160 as a standalone number
// is NOT extracted as a year because it exceeds currentYear+5 (in 2026).
func TestDetector_2160AsYear(t *testing.T) {
	currentYear := time.Now().Year()
	result := extractYear("Movie.2160.mkv")
	if 2160 <= currentYear+5 {
		assert.Equal(t, 2160, result)
	} else {
		assert.Equal(t, 0, result)
	}
}
