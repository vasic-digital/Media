package detector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- matchTVShow: TV pattern matches but extension is non-video, non-empty ---

func TestMatchTVShow_NonVideoExtension(t *testing.T) {
	// A file with a TV show pattern (S01E01) but a non-video extension
	// should NOT match the TV show rule.
	result := matchTVShow("Show.S01E01.txt")
	assert.False(t, result, ".txt is not a video extension")

	result = matchTVShow("Show.S01E01.pdf")
	assert.False(t, result, ".pdf is not a video extension")

	result = matchTVShow("Show.S01E01.jpg")
	assert.False(t, result, ".jpg is not a video extension")
}

func TestMatchTVShow_NoExtension_DirectoryName(t *testing.T) {
	// Directory names (no extension) with TV patterns should match.
	result := matchTVShow("Show S01E01")
	assert.True(t, result, "no extension should match for directory names")
}

func TestMatchTVShow_NoPatternMatch(t *testing.T) {
	// A video file without a TV pattern should not match.
	result := matchTVShow("The.Matrix.1999.mkv")
	assert.False(t, result)
}

// --- extractTVShow: no submatch (pattern matches at index level but
//     FindStringSubmatch returns fewer than 3 groups) ---
//     This is hard to trigger with the existing patterns, but we test
//     the fallback path where no pattern matches at all (returns det
//     with defaults but no name/episode/season).

func TestExtractTVShow_NoPatternMatch(t *testing.T) {
	// extractTVShow is called after matchTVShow, but if the patterns
	// don't match in FindStringSubmatchIndex, it returns a default det.
	det := extractTVShow("random_file_no_pattern.mkv")
	require.NotNil(t, det)
	assert.Equal(t, TypeTVShow, det.Type)
	assert.Equal(t, 0.9, det.Confidence) // base confidence
	assert.Equal(t, 0, det.Season)
	assert.Equal(t, 0, det.Episode)
	assert.Empty(t, det.Name)
}

// --- extractQualityTags: all resolution/source/codec/hdr branches ---

func TestExtractQualityTags_720p(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.720p.WEBRip", tags)
	assert.Equal(t, "720p", tags["resolution"])
	assert.Equal(t, "WEBRip", tags["source"])
}

func TestExtractQualityTags_480p(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.480p.DVDRip", tags)
	assert.Equal(t, "480p", tags["resolution"])
	assert.Equal(t, "DVDRip", tags["source"])
}

func TestExtractQualityTags_WEBDL(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.1080p.WEB-DL.x264", tags)
	assert.Equal(t, "1080p", tags["resolution"])
	assert.Equal(t, "WEB-DL", tags["source"])
	assert.Equal(t, "H.264", tags["video_codec"])
}

func TestExtractQualityTags_WEBDL_NoHyphen(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.WEBDL", tags)
	assert.Equal(t, "WEB-DL", tags["source"])
}

func TestExtractQualityTags_HDTV(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.720p.HDTV", tags)
	assert.Equal(t, "720p", tags["resolution"])
	assert.Equal(t, "HDTV", tags["source"])
}

func TestExtractQualityTags_H264_Variants(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"x264", "Movie.x264"},
		{"h264", "Movie.h264"},
		{"avc", "Movie.AVC"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := make(map[string]string)
			extractQualityTags(tc.input, tags)
			assert.Equal(t, "H.264", tags["video_codec"])
		})
	}
}

func TestExtractQualityTags_AAC(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.1080p.AAC", tags)
	assert.Equal(t, "AAC", tags["audio_codec"])
}

func TestExtractQualityTags_AC3(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.1080p.AC3", tags)
	assert.Equal(t, "AC3", tags["audio_codec"])
}

func TestExtractQualityTags_DolbyVision(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.2160p.Dolby.Vision", tags)
	assert.Equal(t, "true", tags["hdr"])
}

func TestExtractQualityTags_4K_UHD(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.4K.UHD", tags)
	assert.Equal(t, "2160p", tags["resolution"])
}

func TestExtractQualityTags_NoTags(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("simple_movie", tags)
	assert.Empty(t, tags)
}

func TestExtractQualityTags_BrRip(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.BrRip", tags)
	assert.Equal(t, "BluRay", tags["source"])
}

// --- newExtSet ---

func TestNewExtSet_Empty(t *testing.T) {
	s := newExtSet()
	assert.Len(t, s, 0)
}

func TestNewExtSet_Multiple(t *testing.T) {
	s := newExtSet(".a", ".b", ".c")
	assert.Len(t, s, 3)
	_, ok := s[".a"]
	assert.True(t, ok)
}

// --- Detect: multiple custom rules with different confidence ---

func TestDetect_HigherConfidenceRuleWins(t *testing.T) {
	e := NewEngine()

	e.AddRule(Rule{
		Name:     "low_conf",
		Type:     TypeGame,
		Priority: 200,
		Match:    func(f string) bool { return true },
		Extract: func(f string) *Detection {
			return &Detection{
				Type:       TypeGame,
				Confidence: 0.5,
				Name:       "low",
				Tags:       make(map[string]string),
			}
		},
	})

	e.AddRule(Rule{
		Name:     "high_conf",
		Type:     TypeGame,
		Priority: 200,
		Match:    func(f string) bool { return true },
		Extract: func(f string) *Detection {
			return &Detection{
				Type:       TypeGame,
				Confidence: 0.99,
				Name:       "high",
				Tags:       make(map[string]string),
			}
		},
	})

	det := e.Detect("test.rom")
	require.NotNil(t, det)
	assert.Equal(t, "high", det.Name)
	assert.Equal(t, 0.99, det.Confidence)
}

// --- Detect: rule that matches but Extract returns nil ---

func TestDetect_RuleExtractReturnsNil(t *testing.T) {
	e := &Engine{rules: make([]Rule, 0)}

	e.AddRule(Rule{
		Name:     "nil_extract",
		Type:     TypeGame,
		Priority: 200,
		Match:    func(f string) bool { return true },
		Extract:  func(f string) *Detection { return nil },
	})

	det := e.Detect("anything.bin")
	require.NotNil(t, det)
	assert.Equal(t, TypeUnknown, det.Type)
}

// --- extractTVShow with 1x01 pattern (second regex) ---

func TestExtractTVShow_NumberedFormat(t *testing.T) {
	det := extractTVShow("Show.Name.1x05.Title.mkv")
	require.NotNil(t, det)
	assert.Equal(t, TypeTVShow, det.Type)
	assert.Equal(t, 1, det.Season)
	assert.Equal(t, 5, det.Episode)
	assert.Contains(t, det.Name, "Show")
}

// --- extractTVShow with "Season X Episode Y" pattern (third regex) ---

func TestExtractTVShow_SeasonEpisodeWords(t *testing.T) {
	det := extractTVShow("The Show Season 3 Episode 7.mkv")
	require.NotNil(t, det)
	assert.Equal(t, TypeTVShow, det.Type)
	assert.Equal(t, 3, det.Season)
	assert.Equal(t, 7, det.Episode)
	assert.Equal(t, 0.95, det.Confidence)
}

// --- extractMovie: movie without year ---

func TestExtractMovie_NoYear(t *testing.T) {
	det := extractMovie("some_movie.mkv")
	require.NotNil(t, det)
	assert.Equal(t, TypeMovie, det.Type)
	assert.Equal(t, 0, det.Year)
	assert.Equal(t, 0.7, det.Confidence)
}

// --- extractMusic: without artist-title pattern ---

func TestExtractMusic_NoArtistSeparator(t *testing.T) {
	det := extractMusic("track05.mp3")
	require.NotNil(t, det)
	assert.Equal(t, TypeMusic, det.Type)
	assert.Equal(t, 0.85, det.Confidence)
	assert.NotContains(t, det.Tags, "artist")
}

// --- extractBook: without author pattern ---

func TestExtractBook_NoAuthorSeparator(t *testing.T) {
	det := extractBook("MyBook.pdf")
	require.NotNil(t, det)
	assert.Equal(t, TypeBook, det.Type)
	assert.Equal(t, 0.85, det.Confidence)
	assert.NotContains(t, det.Tags, "author")
}
