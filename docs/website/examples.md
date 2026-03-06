# Examples

## Batch Processing with the Pipeline

Process multiple files and filter by media type:

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.media/pkg/detector"
    "digital.vasic.media/pkg/manager"
    "digital.vasic.media/pkg/provider"
)

func main() {
    pipeline := manager.NewPipeline(provider.NewRegistry(), nil)

    files := []string{
        "The.Godfather.1972.1080p.BluRay.mkv",
        "Breaking.Bad.S05E16.Felina.720p.mkv",
        "Pink Floyd - Wish You Were Here.flac",
        "setup.exe",
    }

    results := pipeline.ProcessBatch(context.Background(), files)

    for _, r := range results {
        if r.Detection.Type == detector.TypeMovie {
            fmt.Printf("Movie: %s (%d)\n", r.Metadata.Title, r.Metadata.Year)
        }
    }
}
```

## Custom Detection Rules

Add a custom rule to detect game ROM files:

```go
import "digital.vasic.media/pkg/detector"

engine := detector.NewEngine()

engine.AddRule(detector.Rule{
    Name:     "game_roms",
    Type:     detector.TypeGame,
    Priority: 60,
    Match: func(filename string) bool {
        ext := strings.ToLower(filepath.Ext(filename))
        return ext == ".nes" || ext == ".snes" || ext == ".gba"
    },
    Extract: func(filename string) *detector.Detection {
        base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
        return &detector.Detection{
            Type:       detector.TypeGame,
            Confidence: 0.9,
            Name:       base,
            Extension:  strings.TrimPrefix(filepath.Ext(filename), "."),
            Tags:       map[string]string{"platform": "retro"},
        }
    },
})

det := engine.Detect("Super Mario World.snes")
fmt.Printf("%s: %s (platform: %s)\n", det.Type, det.Name, det.Tags["platform"])
// game: Super Mario World (platform: retro)
```

## Metadata Utilities

Use helper functions for display formatting and title normalization:

```go
import "digital.vasic.media/pkg/metadata"

// Format file sizes
fmt.Println(metadata.FormatFileSize(1536 * 1024 * 1024)) // "1.50 GB"
fmt.Println(metadata.FormatFileSize(750 * 1024))          // "750.00 KB"

// Normalize titles for comparison
a := metadata.NormalizeTitle("The.Dark.Knight.2008")
b := metadata.NormalizeTitle("the dark knight 2008")
fmt.Println(a == b) // true (both: "dark knight 2008")

// Sanitize filenames for cross-platform safety
safe := metadata.SanitizeFilename("My Movie: The Sequel?")
fmt.Println(safe) // "My Movie_ The Sequel_"
```
