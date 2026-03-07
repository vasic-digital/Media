package analyzer

import (
	"sync"
	"testing"

	"digital.vasic.media/pkg/detector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Nil / Zero-Value Inputs ---

func TestAnalyzer_NilInput(t *testing.T) {
	a := NewFilenameAnalyzer()

	t.Run("empty_string_path", func(t *testing.T) {
		_, err := a.Analyze("")
		assert.Error(t, err, "empty path should return error")
	})

	t.Run("dot_only_path", func(t *testing.T) {
		_, err := a.Analyze(".")
		assert.Error(t, err, "dot-only path should return error")
	})

	t.Run("slash_only_path", func(t *testing.T) {
		// filepath.Base("/") returns "/", which is not "" or "."
		meta, err := a.Analyze("/")
		// Depending on OS, "/" may be valid base; just ensure no panic
		if err == nil {
			require.NotNil(t, meta)
			assert.Equal(t, detector.TypeUnknown, meta.MediaType)
		}
	})

	t.Run("whitespace_only_path", func(t *testing.T) {
		meta, err := a.Analyze("   ")
		// filepath.Base("   ") returns "   " which is not empty
		if err == nil {
			require.NotNil(t, meta)
		}
	})
}

// --- Concurrent Analysis ---

func TestAnalyzer_ConcurrentAnalysis(t *testing.T) {
	a := NewFilenameAnalyzer()

	files := []string{
		"/movies/The.Matrix.1999.1080p.mkv",
		"/tv/Breaking.Bad.S01E01.720p.mkv",
		"/music/Pink Floyd - Comfortably Numb.mp3",
		"/books/Isaac Asimov - Foundation.epub",
		"/photos/vacation.jpg",
		"/downloads/installer.exe",
		"/docs/report.docx",
		"/movies/Inception.2010.2160p.BluRay.x265.mkv",
		"/tv/Game.of.Thrones.S08E06.mp4",
		"/music/track01.flac",
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)
	results := make(chan *Metadata, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			path := files[idx%len(files)]
			meta, err := a.Analyze(path)
			if err != nil {
				errors <- err
				return
			}
			results <- meta
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	// Collect errors - none should have occurred for valid files
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	assert.Empty(t, errs, "no errors expected for valid file paths")

	// Collect results - should have goroutines results
	var metas []*Metadata
	for meta := range results {
		metas = append(metas, meta)
	}
	assert.Equal(t, goroutines, len(metas), "all goroutines should produce results")

	// Verify each result has basic fields populated
	for _, meta := range metas {
		assert.NotEmpty(t, meta.Title, "title should not be empty")
		assert.NotNil(t, meta.Tags, "tags should not be nil")
	}
}

// --- Unknown Extension ---

func TestAnalyzer_UnknownExtension(t *testing.T) {
	a := NewFilenameAnalyzer()

	tests := []struct {
		name string
		path string
	}{
		{"xyz_extension", "/files/document.xyz"},
		{"custom_extension", "/files/data.custom"},
		{"numbered_extension", "/files/archive.001"},
		{"long_extension", "/files/file.longextension"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta, err := a.Analyze(tc.path)
			require.NoError(t, err, "unknown extension should not error")
			require.NotNil(t, meta)
			assert.Equal(t, detector.TypeUnknown, meta.MediaType)
			assert.NotEmpty(t, meta.Title)
			assert.NotNil(t, meta.Tags)
		})
	}
}

// --- No Extension ---

func TestAnalyzer_NoExtension(t *testing.T) {
	a := NewFilenameAnalyzer()

	tests := []struct {
		name string
		path string
	}{
		{"simple_name", "/files/README"},
		{"name_with_dots_in_dir", "/some.dir/filename"},
		{"all_caps", "/FILES/MAKEFILE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta, err := a.Analyze(tc.path)
			require.NoError(t, err)
			require.NotNil(t, meta)
			assert.Equal(t, detector.TypeUnknown, meta.MediaType)
			assert.NotEmpty(t, meta.Title)
		})
	}
}

// --- Double Extension ---

func TestAnalyzer_DoubleExtension(t *testing.T) {
	a := NewFilenameAnalyzer()

	tests := []struct {
		name      string
		path      string
		wantType  detector.MediaType
		wantTitle string
	}{
		{
			name:     "tar_gz",
			path:     "/downloads/archive.tar.gz",
			wantType: detector.TypeUnknown,
		},
		{
			name:     "mkv_part",
			path:     "/movies/movie.mkv.part",
			wantType: detector.TypeUnknown, // .part not recognized
		},
		{
			name:     "mp3_bak",
			path:     "/music/song.mp3.bak",
			wantType: detector.TypeUnknown, // .bak not recognized
		},
		{
			name:     "en_srt",
			path:     "/subs/movie.en.srt",
			wantType: detector.TypeUnknown, // .srt not in known extensions
		},
		{
			name:      "real_double_ext_inner_recognized",
			path:      "/movies/movie.2020.mkv",
			wantType:  detector.TypeMovie,
			wantTitle: "movie",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta, err := a.Analyze(tc.path)
			require.NoError(t, err)
			require.NotNil(t, meta)
			assert.Equal(t, tc.wantType, meta.MediaType)
			if tc.wantTitle != "" {
				assert.Equal(t, tc.wantTitle, meta.Title)
			}
		})
	}
}
