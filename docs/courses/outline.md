# Course: Media Detection and Metadata in Go

## Module Overview

This course covers the `digital.vasic.media` module, teaching how to detect media types from filenames, extract structured metadata, integrate external metadata providers, and build a complete detection pipeline. The module handles movies, TV shows, music, books, photos, games, and software.

## Prerequisites

- Intermediate Go knowledge (interfaces, regex, goroutines)
- Familiarity with media file naming conventions
- Go 1.24+ installed

## Lessons

| # | Title | Duration |
|---|-------|----------|
| 1 | Media Type Detection with Rule-Based Engine | 40 min |
| 2 | Filename Analysis and Metadata Extraction | 35 min |
| 3 | External Providers and the Detection Pipeline | 45 min |

## Source Files

- `pkg/detector/` -- Rule-based media type detection engine
- `pkg/analyzer/` -- Filename analysis and metadata extraction
- `pkg/provider/` -- External metadata provider interfaces and registry
- `pkg/metadata/` -- Quality info, resolution, and helper functions
- `pkg/models/` -- Domain model types (MediaItem, MediaFile, etc.)
- `pkg/manager/` -- Detection pipeline facade
