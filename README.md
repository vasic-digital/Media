# digital.vasic.media

A standalone Go module for media file type detection, metadata extraction, and external metadata provider integration.

## Packages

- **`pkg/detector`** - Rule-based media type detection from filenames and extensions. Supports movies, TV shows (S01E02 patterns), music, books, photos, and software.
- **`pkg/analyzer`** - Metadata extraction from file paths. Extracts titles, years, resolution, codecs, season/episode numbers, and artist/author information.
- **`pkg/provider`** - Interface definitions and a concurrent registry for external metadata providers (TMDB, IMDB, MusicBrainz, etc.).
- **`pkg/metadata`** - Shared types (QualityInfo, Resolution, FileInfo) and utility functions (file size formatting, title normalization, filename sanitization).

## Installation

```bash
go get digital.vasic.media
```

## Usage

### Detect media type from a filename

```go
package main

import (
    "fmt"
    "digital.vasic.media/pkg/detector"
)

func main() {
    engine := detector.NewEngine()

    det := engine.Detect("Breaking.Bad.S01E01.720p.BluRay.mkv")
    fmt.Printf("Type: %s, Season: %d, Episode: %d\n", det.Type, det.Season, det.Episode)
    // Output: Type: tv_show, Season: 1, Episode: 1

    det = engine.Detect("The.Matrix.1999.1080p.BluRay.x264.mkv")
    fmt.Printf("Type: %s, Name: %s, Year: %d\n", det.Type, det.Name, det.Year)
    // Output: Type: movie, Name: The Matrix, Year: 1999
}
```

### Analyze media metadata

```go
package main

import (
    "fmt"
    "digital.vasic.media/pkg/analyzer"
)

func main() {
    a := analyzer.NewFilenameAnalyzer()

    meta, _ := a.Analyze("/movies/Inception.2010.2160p.BluRay.x265.mkv")
    fmt.Printf("Title: %s, Year: %d, Resolution: %s, Codec: %s\n",
        meta.Title, meta.Year, meta.Resolution, meta.Codec)
}
```

### Register and search metadata providers

```go
package main

import (
    "context"
    "digital.vasic.media/pkg/provider"
)

func main() {
    registry := provider.NewRegistry()
    // Register your Provider implementations
    // registry.Register(myTMDBProvider)

    results, _ := registry.Search(context.Background(), "The Matrix")
    for _, r := range results {
        fmt.Printf("[%s] %s (%d) - %.1f\n", r.Source, r.Title, r.Year, r.Rating)
    }
}
```

## Development

```bash
# Build
go build ./...

# Test
go test ./... -count=1

# Test with verbose output
go test -v ./... -count=1

# Vet
go vet ./...
```

## Requirements

- Go 1.24.0+
- Test dependency: `github.com/stretchr/testify`

## License

Copyright (c) Milos Vasic. All rights reserved.
