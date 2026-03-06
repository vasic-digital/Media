# Media Architecture

## Purpose

`digital.vasic.media` is a standalone Go module for media file type detection, metadata
extraction, and external metadata enrichment. It provides a rule-based detection engine,
a filename analyzer, a provider registry for external sources (TMDB, IMDB, MusicBrainz),
shared metadata types, and a pipeline manager that orchestrates all three stages.

## Package Overview

| Package | Responsibility |
|---------|---------------|
| `pkg/detector` | Rule-based media type detection from filenames using file extensions, regex patterns, and priority scoring |
| `pkg/analyzer` | Metadata extraction from file paths by wrapping the detection engine and parsing titles, years, codecs, and resolution |
| `pkg/provider` | External metadata provider interface and concurrent fan-out registry for searching TMDB, IMDB, etc. |
| `pkg/metadata` | Shared value types (`QualityInfo`, `Resolution`, `FileInfo`) and utility functions (title normalization, filename sanitization, file size formatting) |
| `pkg/models` | Domain model structs for the full media entity system: `MediaItem`, `MediaFile`, `MediaCollection`, `SearchRequest`, `QualityInfo`, etc. |
| `pkg/manager` | Facade that wires detector, analyzer, and provider registry into a single detection pipeline |

## Design Patterns

| Package | Pattern | Rationale |
|---------|---------|-----------|
| `pkg/detector` | **Rule / Strategy** | Each `Rule` encapsulates a match function and an extract function; rules are prioritized and the highest-confidence match wins |
| `pkg/detector` | **Open/Closed Principle** | New media types are added via `Engine.AddRule()` without modifying existing rules |
| `pkg/analyzer` | **Delegation** | `FilenameAnalyzer` delegates type detection to `detector.Engine` and layers metadata extraction on top |
| `pkg/provider` | **Registry** | Thread-safe map of named providers with dynamic registration and lookup |
| `pkg/provider` | **Fan-Out / Fan-In** | `Registry.Search()` queries all providers concurrently via goroutines and aggregates results |
| `pkg/metadata` | **Value Object** | `QualityInfo`, `Resolution`, and `FileInfo` are immutable data carriers with display methods |
| `pkg/models` | **Domain Model** | Rich struct hierarchy representing the full media entity graph with JSON serialization |
| `pkg/manager` | **Facade** | `Pipeline` hides the three-step workflow (detect, analyze, enrich) behind a single `Process()` call |

## Dependency Diagram

```
  +----------+
  | manager  |  (Facade)
  +----+-----+
       |
       +----------+-----------+
       |          |           |
  +----+-----+ +-+--------+ ++----------+
  | detector | | analyzer | | provider  |
  +----------+ +----+-----+ +-----------+
                    |
                    | uses
               +----+-----+
               | detector  |
               +----------+

  +----------+
  | metadata |   (standalone utility types, no internal deps)
  +----------+

  +----------+
  |  models  |   (standalone domain types, no internal deps)
  +----------+

  analyzer depends on detector.
  manager depends on detector, analyzer, and provider.
  metadata and models are leaf packages with no internal dependencies.
```

## Key Interfaces

```go
// pkg/analyzer -- consumers can implement custom analyzers:
type Analyzer interface {
    Analyze(path string) (*Metadata, error)
    SupportedTypes() []detector.MediaType
}

// pkg/provider -- implemented by TMDB, IMDB, MusicBrainz adapters:
type Provider interface {
    Name() string
    Search(ctx context.Context, query string) ([]*SearchResult, error)
    GetByID(ctx context.Context, id string) (*SearchResult, error)
}

// pkg/manager -- minimal logging interface:
type Logger interface {
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
}
```

### Detection Engine Rules

```go
// pkg/detector -- rules are structs with match/extract functions:
type Rule struct {
    Name     string
    Type     MediaType
    Match    func(filename string) bool
    Extract  func(filename string) *Detection
    Priority int   // higher priority wins when multiple rules match
}
```

Built-in rules and their priorities:

| Rule | Type | Priority |
|------|------|----------|
| `tv_show_pattern` | tv_show | 100 |
| `video_extensions` | movie | 50 |
| `audio_extensions` | music | 50 |
| `image_extensions` | photo | 50 |
| `book_extensions` | book | 50 |
| `software_extensions` | software | 50 |

### Detection Pipeline (manager.Pipeline)

```
  Path
   |
   v
  [1] detector.Engine.Detect()  --> Detection{Type, Confidence, Name, Year, ...}
   |
   v
  [2] analyzer.Analyze()        --> Metadata{Title, Year, Resolution, Codec, Tags}
   |
   v
  [3] provider.Registry.Search() --> []*SearchResult (from TMDB, IMDB, etc.)
   |
   v
  Result{Path, Detection, Metadata, Providers, Error}
```

## Usage Example

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
    // Set up a provider registry (add TMDB, IMDB implementations as needed).
    registry := provider.NewRegistry()
    // registry.Register(tmdbProvider)

    // Create the pipeline.
    pipeline := manager.NewPipeline(registry, nil)

    // Process a single file.
    result := pipeline.Process(context.Background(), "/media/Breaking.Bad.S01E01.720p.mkv")

    fmt.Printf("Type:       %s\n", result.Detection.Type)       // tv_show
    fmt.Printf("Confidence: %.2f\n", result.Detection.Confidence) // 0.95
    fmt.Printf("Title:      %s\n", result.Metadata.Title)         // Breaking Bad
    fmt.Printf("Season:     %s\n", result.Metadata.Tags["season"])  // 1
    fmt.Printf("Episode:    %s\n", result.Metadata.Tags["episode"]) // 1

    // Add a custom detection rule.
    pipeline.Engine().AddRule(detector.Rule{
        Name:     "iso_files",
        Type:     detector.TypeSoftware,
        Priority: 60,
        Match:    func(f string) bool { return filepath.Ext(f) == ".iso" },
        Extract:  func(f string) *detector.Detection {
            return &detector.Detection{Type: detector.TypeSoftware, Confidence: 0.9, Name: f}
        },
    })

    // Batch processing.
    paths := []string{"/media/movie.mkv", "/music/song.flac", "/books/novel.epub"}
    results := pipeline.ProcessBatch(context.Background(), paths)
    for _, r := range results {
        fmt.Printf("%s -> %s (%.0f%%)\n", r.Path, r.Detection.Type, r.Detection.Confidence*100)
    }
}
```
