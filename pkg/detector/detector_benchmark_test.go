package detector

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkDetect_SingleFile(b *testing.B) {
	e := NewEngine()
	filename := "The.Matrix.1999.1080p.BluRay.x264.DTS.mkv"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Detect(filename)
	}
}

func BenchmarkDetect_TVShow(b *testing.B) {
	e := NewEngine()
	filename := "Breaking.Bad.S01E01.Pilot.720p.BluRay.x264.mkv"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Detect(filename)
	}
}

func BenchmarkDetect_Music(b *testing.B) {
	e := NewEngine()
	filename := "Pink Floyd - Comfortably Numb.flac"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Detect(filename)
	}
}

func BenchmarkDetect_Unknown(b *testing.B) {
	e := NewEngine()
	filename := "random_document.docx"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Detect(filename)
	}
}

func BenchmarkDetect_1000Files(b *testing.B) {
	e := NewEngine()

	// Generate a realistic batch of 1000 filenames
	files := make([]string, 1000)
	extensions := []string{".mkv", ".mp4", ".avi", ".mp3", ".flac", ".jpg", ".pdf", ".exe", ".docx", ".txt"}
	for i := 0; i < 1000; i++ {
		ext := extensions[i%len(extensions)]
		switch {
		case i%10 < 3: // 30% movies
			files[i] = fmt.Sprintf("Movie.Title.%d.%dp.BluRay.x264%s", 1990+i%35, 720+(i%3)*360, ext)
		case i%10 < 5: // 20% TV shows
			files[i] = fmt.Sprintf("Show.Name.S%02dE%02d.720p%s", 1+i%10, 1+i%24, ext)
		case i%10 < 7: // 20% music
			files[i] = fmt.Sprintf("Artist %d - Track %d%s", i%50, i%20, ext)
		default: // 30% mixed
			files[i] = fmt.Sprintf("file_%d%s", i, ext)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			_ = e.Detect(f)
		}
	}
}

func BenchmarkDetect_WithAllRules(b *testing.B) {
	e := NewEngine()

	// Add many custom rules to stress the rule sorting and matching
	for i := 0; i < 50; i++ {
		ext := fmt.Sprintf(".custom%d", i)
		e.AddRule(Rule{
			Name:     fmt.Sprintf("custom_rule_%d", i),
			Type:     TypeGame,
			Priority: 30 + i,
			Match: func(filename string) bool {
				return strings.ToLower(filepath.Ext(filename)) == ext
			},
			Extract: func(filename string) *Detection {
				return &Detection{
					Type:       TypeGame,
					Confidence: 0.8,
					Name:       filenameWithoutExt(filename),
					Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
					Tags:       map[string]string{"source": "custom"},
				}
			},
		})
	}

	filename := "The.Matrix.1999.1080p.BluRay.x264.mkv"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Detect(filename)
	}
}

func BenchmarkDetect_LongFilename(b *testing.B) {
	e := NewEngine()
	filename := strings.Repeat("Very.Long.Title.", 30) + "2020.1080p.BluRay.mkv"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Detect(filename)
	}
}
