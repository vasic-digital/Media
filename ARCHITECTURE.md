# Architecture -- Media

## Purpose

Standalone Go module for media file type detection, metadata extraction, and external metadata provider integration. Provides a rule-based detection engine, filename analysis, a concurrent provider registry, and shared metadata types.

## Structure

```
pkg/
  detector/   Rule-based media type detection from filenames and extensions (movies, TV shows, music, books, photos, software)
  analyzer/   Metadata extraction from file paths: titles, years, resolution, codecs, season/episode, artist/author
  provider/   Interface definitions and concurrent registry for external metadata providers (TMDB, IMDB, MusicBrainz, etc.)
  metadata/   Shared types (QualityInfo, Resolution, FileInfo) and utilities (file size formatting, title normalization, filename sanitization)
```

## Key Components

- **`detector.Engine`** -- Rule-based detection with priorities: TV Show patterns (priority 100), extension-based rules (priority 50), custom rules via AddRule
- **`analyzer.FilenameAnalyzer`** -- Wraps detection engine; extracts structured metadata (title, year, resolution, codec, artist, season, episode)
- **`provider.Registry`** -- Thread-safe registration and lookup of metadata providers; concurrent Search across all providers with graceful error handling
- **`provider.Provider`** -- Interface: Name, Search(ctx, query), GetByID(ctx, id)
- **`metadata.QualityInfo`** -- Quality attributes with comparison and display methods
- **`metadata.Resolution`** -- Human-readable resolution formatting

## Data Flow

```
detector.Engine.Detect(filename) -> iterate rules by priority
    TV Show patterns (S01E02, 1x01) -> Detection{Type: "tv_show", Season, Episode}
    Extension rules (video, audio)  -> Detection{Type: "movie" or "music"}

analyzer.FilenameAnalyzer.Analyze(path) -> Engine.Detect() + regex extraction
    -> AnalysisResult{Title, Year, Resolution, Codec, Season, Episode, Artist}

provider.Registry.Search(ctx, query) -> concurrent search across all registered providers
    -> []SearchResult{Source, Title, Year, Rating}
```

## Dependencies

- `github.com/stretchr/testify` -- Test assertions (only dependency)

## Testing Strategy

Table-driven tests with `testify`. Tests cover TV show pattern detection with various formats, extension-based classification, filename analysis with complex naming schemes, concurrent provider registry search, and metadata type comparison.
