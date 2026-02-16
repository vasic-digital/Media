package detector

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine()
	require.NotNil(t, e)
	assert.True(t, len(e.rules) > 0, "engine should have default rules")
}

// --- TV Show Detection ---

func TestDetect_TVShow_S01E02(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		filename string
		name     string
		season   int
		episode  int
	}{
		{"Breaking.Bad.S01E01.720p.BluRay.x264.mkv", "Breaking Bad", 1, 1},
		{"The.Office.S03E12.1080p.WEB-DL.mkv", "The Office", 3, 12},
		{"Game.of.Thrones.S08E06.The.Iron.Throne.2160p.mkv", "Game of Thrones", 8, 6},
		{"stranger.things.s04e09.webm", "stranger things", 4, 9},
		{"Friends.S10E17.The.Last.One.Part.1.mp4", "Friends", 10, 17},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, TypeTVShow, det.Type)
			assert.Equal(t, tc.season, det.Season)
			assert.Equal(t, tc.episode, det.Episode)
			assert.True(t, det.Confidence >= 0.9)
			assert.NotEmpty(t, det.Name)
		})
	}
}

func TestDetect_TVShow_NumberedFormat(t *testing.T) {
	e := NewEngine()

	det := e.Detect("Lost.1x01.Pilot.Part.1.mkv")
	require.NotNil(t, det)
	assert.Equal(t, TypeTVShow, det.Type)
	assert.Equal(t, 1, det.Season)
	assert.Equal(t, 1, det.Episode)
}

func TestDetect_TVShow_SeasonEpisodeWords(t *testing.T) {
	e := NewEngine()

	det := e.Detect("The Simpsons Season 5 Episode 12.mp4")
	require.NotNil(t, det)
	assert.Equal(t, TypeTVShow, det.Type)
	assert.Equal(t, 5, det.Season)
	assert.Equal(t, 12, det.Episode)
}

// --- Movie Detection ---

func TestDetect_Movie(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		filename string
		year     int
	}{
		{"The.Matrix.1999.1080p.BluRay.x264.mkv", 1999},
		{"Inception.2010.720p.BrRip.mp4", 2010},
		{"Interstellar.(2014).2160p.UHD.mkv", 2014},
		{"Blade.Runner.2049.2017.WEB-DL.avi", 2017},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, TypeMovie, det.Type)
			assert.Equal(t, tc.year, det.Year)
			assert.True(t, det.Confidence >= 0.7)
			assert.NotEmpty(t, det.Name)
		})
	}
}

func TestDetect_MovieWithoutYear(t *testing.T) {
	e := NewEngine()

	det := e.Detect("Pulp.Fiction.1080p.BluRay.mkv")
	require.NotNil(t, det)
	assert.Equal(t, TypeMovie, det.Type)
	assert.Equal(t, 0, det.Year)
	assert.Equal(t, 0.7, det.Confidence)
}

func TestDetect_MovieQualityTags(t *testing.T) {
	e := NewEngine()

	det := e.Detect("The.Matrix.1999.2160p.BluRay.x265.HDR.DTS.mkv")
	require.NotNil(t, det)
	assert.Equal(t, TypeMovie, det.Type)
	assert.Equal(t, "2160p", det.Tags["resolution"])
	assert.Equal(t, "BluRay", det.Tags["source"])
	assert.Equal(t, "H.265", det.Tags["video_codec"])
	assert.Equal(t, "DTS", det.Tags["audio_codec"])
	assert.Equal(t, "true", det.Tags["hdr"])
}

// --- Music Detection ---

func TestDetect_Music(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		filename  string
		mediaType MediaType
	}{
		{"Pink Floyd - Comfortably Numb.mp3", TypeMusic},
		{"01 - Bohemian Rhapsody.flac", TypeMusic},
		{"track05.wav", TypeMusic},
		{"song.aac", TypeMusic},
		{"podcast_episode.ogg", TypeMusic},
		{"windows_media.wma", TypeMusic},
		{"apple_audio.m4a", TypeMusic},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, tc.mediaType, det.Type)
			assert.True(t, det.Confidence >= 0.85)
		})
	}
}

func TestDetect_MusicArtistTitle(t *testing.T) {
	e := NewEngine()

	det := e.Detect("Pink Floyd - Comfortably Numb.mp3")
	require.NotNil(t, det)
	assert.Equal(t, TypeMusic, det.Type)
	assert.Equal(t, "Comfortably Numb", det.Name)
	assert.Equal(t, "Pink Floyd", det.Tags["artist"])
	assert.Equal(t, 0.9, det.Confidence)
}

// --- Photo Detection ---

func TestDetect_Photo(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		filename string
		ext      string
	}{
		{"vacation.jpg", "jpg"},
		{"photo.jpeg", "jpeg"},
		{"screenshot.png", "png"},
		{"animation.gif", "gif"},
		{"bitmap.bmp", "bmp"},
		{"modern.webp", "webp"},
		{"vector.svg", "svg"},
		{"scan.tiff", "tiff"},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, TypePhoto, det.Type)
			assert.Equal(t, tc.ext, det.Extension)
			assert.True(t, det.Confidence >= 0.85)
		})
	}
}

// --- Book Detection ---

func TestDetect_Book(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		filename string
		ext      string
	}{
		{"The Great Gatsby.pdf", "pdf"},
		{"novel.epub", "epub"},
		{"ebook.mobi", "mobi"},
		{"kindle_book.azw3", "azw3"},
		{"comic.cbr", "cbr"},
		{"comic_archive.cbz", "cbz"},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, TypeBook, det.Type)
			assert.Equal(t, tc.ext, det.Extension)
			assert.True(t, det.Confidence >= 0.85)
		})
	}
}

func TestDetect_BookAuthorTitle(t *testing.T) {
	e := NewEngine()

	det := e.Detect("Isaac Asimov - Foundation.epub")
	require.NotNil(t, det)
	assert.Equal(t, TypeBook, det.Type)
	assert.Equal(t, "Foundation", det.Name)
	assert.Equal(t, "Isaac Asimov", det.Tags["author"])
}

// --- Software Detection ---

func TestDetect_Software(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		filename string
		ext      string
	}{
		{"installer.exe", "exe"},
		{"setup.msi", "msi"},
		{"app.dmg", "dmg"},
		{"package.deb", "deb"},
		{"package.rpm", "rpm"},
		{"MyApp.AppImage", "AppImage"},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			det := e.Detect(tc.filename)
			require.NotNil(t, det)
			assert.Equal(t, TypeSoftware, det.Type)
			assert.Equal(t, tc.ext, det.Extension)
			assert.True(t, det.Confidence >= 0.85)
		})
	}
}

// --- Unknown Type ---

func TestDetect_Unknown(t *testing.T) {
	e := NewEngine()

	det := e.Detect("document.docx")
	require.NotNil(t, det)
	assert.Equal(t, TypeUnknown, det.Type)
	assert.Equal(t, 0.0, det.Confidence)
	assert.Equal(t, "document", det.Name)
	assert.Equal(t, "docx", det.Extension)
}

// --- Custom Rules ---

func TestAddRule_CustomDetection(t *testing.T) {
	e := NewEngine()

	e.AddRule(Rule{
		Name:     "game_rom",
		Type:     TypeGame,
		Priority: 60,
		Match: func(filename string) bool {
			ext := strings.ToLower(filepath.Ext(filename))
			return ext == ".rom" || ext == ".nes" || ext == ".snes"
		},
		Extract: func(filename string) *Detection {
			return &Detection{
				Type:       TypeGame,
				Confidence: 0.9,
				Name:       filenameWithoutExt(filename),
				Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
				Tags:       map[string]string{"platform": "retro"},
			}
		},
	})

	det := e.Detect("super_mario.nes")
	require.NotNil(t, det)
	assert.Equal(t, TypeGame, det.Type)
	assert.Equal(t, "super_mario", det.Name)
	assert.Equal(t, "retro", det.Tags["platform"])
}

// --- TV Show takes priority over Movie for video with S##E## pattern ---

func TestDetect_TVShowPriorityOverMovie(t *testing.T) {
	e := NewEngine()

	det := e.Detect("House.of.Cards.S01E01.720p.BluRay.mkv")
	require.NotNil(t, det)
	// Should detect as TV show, not movie, because of S01E01 pattern
	assert.Equal(t, TypeTVShow, det.Type)
	assert.Equal(t, 1, det.Season)
	assert.Equal(t, 1, det.Episode)
}

// --- Helper function tests ---

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"The.Matrix.1999.1080p.BluRay.x264", "The Matrix"},
		{"Inception.2010.720p.BrRip", "Inception"},
		{"Simple_Title", "Simple Title"},
		{"Multi---Dash", "Multi Dash"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := cleanTitle(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractYear(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"The Matrix (1999)", 1999},
		{"Movie [2020]", 2020},
		{"Film.2015.1080p", 2015},
		{"No year here", 0},
		{"Ancient 1800", 0}, // too old
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := extractYear(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFilenameWithoutExt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"movie.mkv", "movie"},
		{"no_extension", "no_extension"},
		{"/path/to/file.mp4", "file"},
		{"dotted.name.avi", "dotted.name"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := filenameWithoutExt(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractQualityTags(t *testing.T) {
	tags := make(map[string]string)
	extractQualityTags("Movie.2160p.BluRay.x265.HDR.DTS", tags)

	assert.Equal(t, "2160p", tags["resolution"])
	assert.Equal(t, "BluRay", tags["source"])
	assert.Equal(t, "H.265", tags["video_codec"])
	assert.Equal(t, "DTS", tags["audio_codec"])
	assert.Equal(t, "true", tags["hdr"])
}

func TestDetect_CaseInsensitive(t *testing.T) {
	e := NewEngine()

	det := e.Detect("MOVIE.MKV")
	require.NotNil(t, det)
	assert.Equal(t, TypeMovie, det.Type)

	det = e.Detect("song.MP3")
	require.NotNil(t, det)
	assert.Equal(t, TypeMusic, det.Type)

	det = e.Detect("image.PNG")
	require.NotNil(t, det)
	assert.Equal(t, TypePhoto, det.Type)
}

func TestDetect_AllVideoExtensions(t *testing.T) {
	e := NewEngine()

	exts := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm"}
	for _, ext := range exts {
		t.Run(ext, func(t *testing.T) {
			det := e.Detect("file" + ext)
			require.NotNil(t, det)
			// Could be movie or TV show depending on name, but should not be unknown
			assert.NotEqual(t, TypeUnknown, det.Type)
		})
	}
}

func TestDetect_AllAudioExtensions(t *testing.T) {
	e := NewEngine()

	exts := []string{".mp3", ".flac", ".wav", ".aac", ".ogg", ".wma", ".m4a"}
	for _, ext := range exts {
		t.Run(ext, func(t *testing.T) {
			det := e.Detect("file" + ext)
			require.NotNil(t, det)
			assert.Equal(t, TypeMusic, det.Type)
		})
	}
}

func TestDetect_AllImageExtensions(t *testing.T) {
	e := NewEngine()

	exts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".tiff"}
	for _, ext := range exts {
		t.Run(ext, func(t *testing.T) {
			det := e.Detect("file" + ext)
			require.NotNil(t, det)
			assert.Equal(t, TypePhoto, det.Type)
		})
	}
}

func TestDetect_AllBookExtensions(t *testing.T) {
	e := NewEngine()

	exts := []string{".pdf", ".epub", ".mobi", ".azw3", ".cbr", ".cbz"}
	for _, ext := range exts {
		t.Run(ext, func(t *testing.T) {
			det := e.Detect("file" + ext)
			require.NotNil(t, det)
			assert.Equal(t, TypeBook, det.Type)
		})
	}
}

func TestDetect_AllSoftwareExtensions(t *testing.T) {
	e := NewEngine()

	exts := []string{".exe", ".msi", ".dmg", ".deb", ".rpm", ".appimage"}
	for _, ext := range exts {
		t.Run(ext, func(t *testing.T) {
			det := e.Detect("file" + ext)
			require.NotNil(t, det)
			assert.Equal(t, TypeSoftware, det.Type)
		})
	}
}
