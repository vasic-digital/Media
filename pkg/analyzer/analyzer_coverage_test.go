package analyzer

import (
	"testing"

	"digital.vasic.media/pkg/detector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Analyze: TV show without resolution/codec in tags (falls back to extractResolution/extractCodec) ---

func TestAnalyze_TVShow_FallbackResolutionCodec(t *testing.T) {
	a := NewFilenameAnalyzer()

	// A TV show file without resolution or codec in the filename —
	// the tags won't have "resolution" or "video_codec", so the code
	// falls through to extractResolution and extractCodec.
	meta, err := a.Analyze("/tv/Show.S02E03.mkv")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeTVShow, meta.MediaType)
	assert.Equal(t, "", meta.Resolution) // no resolution info
	assert.Equal(t, "", meta.Codec)      // no codec info
	assert.Equal(t, "2", meta.Tags["season"])
	assert.Equal(t, "3", meta.Tags["episode"])
}

// --- Analyze: movie with resolution in tags (via detection quality tags) ---

func TestAnalyze_Movie_ResolutionFromTags(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/movies/Film.2020.2160p.BluRay.x265.mkv")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeMovie, meta.MediaType)
	assert.Equal(t, "2160p", meta.Resolution)
	assert.Equal(t, "H.265", meta.Codec)
}

// --- Analyze: movie without resolution/codec (falls back to extractResolution/extractCodec which return "") ---

func TestAnalyze_Movie_NoResolutionNoCodec(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/movies/simple_movie.mkv")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeMovie, meta.MediaType)
	assert.Equal(t, "", meta.Resolution)
	assert.Equal(t, "", meta.Codec)
}

// --- Analyze: music without artist tag ---

func TestAnalyze_Music_NoArtist(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/music/track05.mp3")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeMusic, meta.MediaType)
	_, hasArtist := meta.Tags["artist"]
	assert.False(t, hasArtist, "no artist tag expected for simple filenames")
}

// --- Analyze: book without author tag ---

func TestAnalyze_Book_NoAuthor(t *testing.T) {
	a := NewFilenameAnalyzer()

	meta, err := a.Analyze("/books/MyBook.pdf")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, detector.TypeBook, meta.MediaType)
	_, hasAuthor := meta.Tags["author"]
	assert.False(t, hasAuthor, "no author tag expected for simple filenames")
}

// --- Analyze: dot-only path (invalid) ---

func TestAnalyze_DotPath(t *testing.T) {
	a := NewFilenameAnalyzer()

	_, err := a.Analyze(".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

// --- extractResolution: case insensitivity ---

func TestExtractResolution_CaseInsensitive(t *testing.T) {
	assert.Equal(t, "2160p", extractResolution("MOVIE.2160P.MKV"))
	assert.Equal(t, "2160p", extractResolution("movie.4K.mkv"))
	assert.Equal(t, "2160p", extractResolution("movie.UHD.mkv"))
	assert.Equal(t, "1080p", extractResolution("movie.1080P.mkv"))
}

// --- extractCodec: h.265 with dot ---

func TestExtractCodec_H265WithDot(t *testing.T) {
	result := extractCodec("movie.h.265.mkv")
	assert.Equal(t, "H.265", result)
}

func TestExtractCodec_H264WithDot(t *testing.T) {
	result := extractCodec("movie.h.264.mkv")
	assert.Equal(t, "H.264", result)
}
