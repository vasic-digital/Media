# Media Module

`digital.vasic.media` is a standalone Go module for media file type detection, metadata extraction, external metadata provider interfaces, and a unified detection pipeline. It detects movies, TV shows, music, books, photos, games, and software from filenames, extracts structured metadata, and queries external providers for enrichment.

## Key Features

- **Media type detection** -- Rule-based engine that identifies movies, TV shows, music, books, photos, games, and software from filenames and extensions
- **Filename analysis** -- Extracts title, year, resolution, codec, season/episode, artist/author from filename patterns
- **Provider registry** -- Thread-safe registry for external metadata providers (TMDB, IMDB, etc.) with concurrent search
- **Metadata utilities** -- Quality comparison, resolution formatting, file size formatting, title normalization, filename sanitization
- **Domain models** -- Complete media entity types including MediaItem, MediaFile, MediaCollection, QualityInfo, and SearchRequest
- **Detection pipeline** -- Facade that orchestrates detection, analysis, and provider enrichment in a single call

## Package Overview

| Package | Purpose |
|---------|---------|
| `pkg/detector` | Media type detection from filenames and extensions |
| `pkg/analyzer` | Metadata extraction and analysis from file paths |
| `pkg/provider` | External metadata provider interfaces and registry |
| `pkg/metadata` | Quality info, resolution, file size formatting, title normalization |
| `pkg/models` | Domain model types (MediaItem, MediaFile, MediaCollection, etc.) |
| `pkg/manager` | Detection pipeline facade (detect, analyze, enrich) |

## Installation

```bash
go get digital.vasic.media
```

Requires Go 1.24 or later. Only external dependency is `testify` for tests.
