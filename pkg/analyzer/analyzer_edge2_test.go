package analyzer_test

import (
	"strings"
	"testing"

	"digital.vasic.media/pkg/analyzer"
	"digital.vasic.media/pkg/detector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Zero-Byte / Empty Filenames ---

func TestFilenameAnalyzer_EmptyPath(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	tests := []struct {
		name string
		path string
	}{
		{"empty_string", ""},
		{"single_dot", "."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := a.Analyze(tc.path)
			assert.Error(t, err, "expected error for path %q", tc.path)
		})
	}
}

// --- Extremely Long Filenames (255+ chars) ---

func TestFilenameAnalyzer_ExtremelyLongFilename(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	// 300 character filename with video extension
	longName := strings.Repeat("A", 300) + ".2020.1080p.mkv"
	meta, err := a.Analyze(longName)
	require.NoError(t, err)
	assert.Equal(t, detector.TypeMovie, meta.MediaType)
	assert.Equal(t, 2020, meta.Year)
	assert.Equal(t, "1080p", meta.Resolution)
	assert.NotEmpty(t, meta.Title)

	// 500 character filename without known extension
	longUnknown := strings.Repeat("X", 500) + ".unknownext"
	meta2, err := a.Analyze(longUnknown)
	require.NoError(t, err)
	assert.Equal(t, detector.TypeUnknown, meta2.MediaType)
}

// --- Filenames With Unicode ---

func TestFilenameAnalyzer_UnicodeFilenames(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	tests := []struct {
		name      string
		path      string
		wantType  detector.MediaType
	}{
		{
			"japanese_movie",
			"/movies/\u5343\u3068\u5343\u5c0b\u306e\u795e\u96a0\u3057.2001.mkv",
			detector.TypeMovie,
		},
		{
			"cyrillic_music",
			"/music/\u041c\u043e\u0441\u043a\u0432\u0430 - \u041d\u043e\u0447\u044c.mp3",
			detector.TypeMusic,
		},
		{
			"korean_tv",
			"\uc624\uc9d5\uc5b4\uac8c\uc784.S01E03.mkv",
			detector.TypeTVShow,
		},
		{
			"emoji_filename",
			"\U0001f3ac\U0001f3a5movie.mp4",
			detector.TypeMovie,
		},
		{
			"arabic_book",
			"\u0643\u062a\u0627\u0628 - \u0639\u0631\u0628\u064a.pdf",
			detector.TypeBook,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			meta, err := a.Analyze(tc.path)
			require.NoError(t, err)
			assert.Equal(t, tc.wantType, meta.MediaType)
		})
	}
}

// --- Unsupported File Extensions ---

func TestFilenameAnalyzer_UnsupportedExtensions(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	tests := []string{
		"file.xyz",
		"document.docx",
		"archive.tar.gz",
		"data.csv",
		"binary.bin",
		"config.yaml",
		"readme.md",
	}

	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			t.Parallel()
			meta, err := a.Analyze(filename)
			require.NoError(t, err)
			assert.Equal(t, detector.TypeUnknown, meta.MediaType)
		})
	}
}

// --- Directory-Like Paths as Input ---

func TestFilenameAnalyzer_DirectoryAsInput(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	// filepath.Base of a directory path without trailing slash
	// returns the last component, which has no extension
	meta, err := a.Analyze("/movies/The Matrix (1999)")
	require.NoError(t, err)
	assert.Equal(t, detector.TypeUnknown, meta.MediaType)
	assert.NotEmpty(t, meta.Title)
}

// --- Filenames With Only Extension ---

func TestFilenameAnalyzer_OnlyExtension(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	tests := []struct {
		name     string
		path     string
		wantType detector.MediaType
	}{
		{"hidden_mkv", ".mkv", detector.TypeMovie},
		{"hidden_mp3", ".mp3", detector.TypeMusic},
		{"hidden_pdf", ".pdf", detector.TypeBook},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			meta, err := a.Analyze(tc.path)
			require.NoError(t, err)
			assert.Equal(t, tc.wantType, meta.MediaType)
		})
	}
}

// --- Filenames With Multiple Dots ---

func TestFilenameAnalyzer_MultipleDots(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	meta, err := a.Analyze("Movie.Name.2020.1080p.BluRay.x264.DTS-GROUP.mkv")
	require.NoError(t, err)
	assert.Equal(t, detector.TypeMovie, meta.MediaType)
	assert.Equal(t, 2020, meta.Year)
}

// --- SupportedTypes ---

func TestFilenameAnalyzer_SupportedTypes(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()
	types := a.SupportedTypes()

	assert.NotEmpty(t, types)
	assert.Contains(t, types, detector.TypeMovie)
	assert.Contains(t, types, detector.TypeTVShow)
	assert.Contains(t, types, detector.TypeMusic)
	assert.Contains(t, types, detector.TypeBook)
}

// --- Zero-Byte File (just extension, no meaningful content in name) ---

func TestFilenameAnalyzer_NoMeaningfulName(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	// Filename is just spaces followed by extension
	meta, err := a.Analyze("   .mp4")
	require.NoError(t, err)
	assert.Equal(t, detector.TypeMovie, meta.MediaType)
}

// --- Codec Extraction Edge Cases ---

func TestFilenameAnalyzer_CodecExtraction(t *testing.T) {
	t.Parallel()

	a := analyzer.NewFilenameAnalyzer()

	tests := []struct {
		name      string
		filename  string
		wantCodec string
	}{
		{"x265", "Movie.2020.x265.mkv", "H.265"},
		{"hevc", "Movie.HEVC.2020.mkv", "H.265"},
		{"h264", "Movie.H264.mkv", "H.264"},
		{"x264", "Movie.x264.mkv", "H.264"},
		{"no_codec", "Movie.2020.mkv", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			meta, err := a.Analyze(tc.filename)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCodec, meta.Codec)
		})
	}
}
