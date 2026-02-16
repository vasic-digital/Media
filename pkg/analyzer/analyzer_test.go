package analyzer

import (
	"testing"

	"digital.vasic.media/pkg/detector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFilenameAnalyzer(t *testing.T) {
	a := NewFilenameAnalyzer()
	require.NotNil(t, a)
	require.NotNil(t, a.engine)
}

func TestFilenameAnalyzer_SupportedTypes(t *testing.T) {
	a := NewFilenameAnalyzer()
	types := a.SupportedTypes()
	assert.Contains(t, types, detector.TypeMovie)
	assert.Contains(t, types, detector.TypeTVShow)
	assert.Contains(t, types, detector.TypeMusic)
	assert.Contains(t, types, detector.TypeBook)
	assert.Contains(t, types, detector.TypePhoto)
	assert.Contains(t, types, detector.TypeSoftware)
}

func TestAnalyze_Movie(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/movies/The.Matrix.1999.1080p.BluRay.x264.mkv")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeMovie, meta.MediaType)
	assert.Equal(t, "The Matrix", meta.Title)
	assert.Equal(t, 1999, meta.Year)
	assert.Equal(t, "1080p", meta.Resolution)
	assert.Equal(t, "H.264", meta.Codec)
}

func TestAnalyze_TVShow(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/tv/Breaking.Bad.S01E01.720p.BluRay.mkv")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeTVShow, meta.MediaType)
	assert.NotEmpty(t, meta.Title)
	assert.Equal(t, "1", meta.Tags["season"])
	assert.Equal(t, "1", meta.Tags["episode"])
	assert.Equal(t, "720p", meta.Resolution)
}

func TestAnalyze_Music(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/music/Pink Floyd - Comfortably Numb.mp3")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeMusic, meta.MediaType)
	assert.Equal(t, "Comfortably Numb", meta.Title)
	assert.Equal(t, "Pink Floyd", meta.Tags["artist"])
}

func TestAnalyze_Book(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/books/Isaac Asimov - Foundation.epub")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeBook, meta.MediaType)
	assert.Equal(t, "Foundation", meta.Title)
	assert.Equal(t, "Isaac Asimov", meta.Tags["author"])
}

func TestAnalyze_Photo(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/photos/vacation.jpg")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypePhoto, meta.MediaType)
	assert.Equal(t, "vacation", meta.Title)
}

func TestAnalyze_Software(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/downloads/installer.exe")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeSoftware, meta.MediaType)
	assert.Equal(t, "installer", meta.Title)
}

func TestAnalyze_InvalidPath(t *testing.T) {
	a := NewFilenameAnalyzer()

	_, err := a.Analyze("")
	assert.Error(t, err)
}

func TestAnalyze_UnknownType(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/docs/report.docx")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeUnknown, meta.MediaType)
}

func TestAnalyze_MovieWithResolution(t *testing.T) {
	a := NewFilenameAnalyzer()

	tests := []struct {
		path       string
		resolution string
	}{
		{"/movies/film.2160p.mkv", "2160p"},
		{"/movies/film.4k.mkv", "2160p"},
		{"/movies/film.1080p.mkv", "1080p"},
		{"/movies/film.720p.mkv", "720p"},
		{"/movies/film.480p.mkv", "480p"},
	}

	for _, tc := range tests {
		t.Run(tc.resolution, func(t *testing.T) {
			meta, err := a.Analyze(tc.path)
			require.NoError(t, err)
			assert.Equal(t, tc.resolution, meta.Resolution)
		})
	}
}

func TestAnalyze_MovieWithCodec(t *testing.T) {
	a := NewFilenameAnalyzer()

	tests := []struct {
		path  string
		codec string
	}{
		{"/movies/film.x265.mkv", "H.265"},
		{"/movies/film.HEVC.mkv", "H.265"},
		{"/movies/film.x264.mkv", "H.264"},
		{"/movies/film.AVC.mkv", "H.264"},
	}

	for _, tc := range tests {
		t.Run(tc.codec, func(t *testing.T) {
			meta, err := a.Analyze(tc.path)
			require.NoError(t, err)
			assert.Equal(t, tc.codec, meta.Codec)
		})
	}
}

func TestExtractResolution(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"movie.2160p.mkv", "2160p"},
		{"movie.4k.mkv", "2160p"},
		{"movie.UHD.mkv", "2160p"},
		{"movie.1080p.mkv", "1080p"},
		{"movie.720p.mkv", "720p"},
		{"movie.480p.mkv", "480p"},
		{"movie.mkv", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := extractResolution(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractCodec(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"movie.x265.mkv", "H.265"},
		{"movie.h265.mkv", "H.265"},
		{"movie.HEVC.mkv", "H.265"},
		{"movie.x264.mkv", "H.264"},
		{"movie.h264.mkv", "H.264"},
		{"movie.AVC.mkv", "H.264"},
		{"movie.AV1.mkv", "AV1"},
		{"movie.VP9.mkv", "VP9"},
		{"movie.mkv", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := extractCodec(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
