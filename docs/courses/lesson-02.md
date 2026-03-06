# Lesson 2: Filename Analysis and Metadata Extraction

## Learning Objectives

- Extract structured metadata from filenames using the analyzer
- Use the metadata package for quality comparison and formatting
- Work with domain model types for representing media entities

## Key Concepts

- **FilenameAnalyzer**: Wraps the detection engine and produces `Metadata` structs with title, year, resolution, codec, genre, and tags. It delegates detection to the engine and enriches the result with additional pattern matching.
- **Quality Comparison**: `QualityInfo.IsBetterThan(other)` compares quality scores. `Resolution.String()` formats dimensions into human-readable labels (4K/UHD, 1080p, 720p, 480p).
- **Domain Models**: `models.MediaItem` represents a detected entity with hierarchy support (parent_id for TV show seasons/episodes). `models.MediaFile` links physical files to entities. `models.SearchRequest` parameterizes media queries.

## Code Walkthrough

### Source: `pkg/analyzer/analyzer.go`

The `Analyze` method calls the detection engine, copies tags, extracts resolution and codec from tags or filename patterns, and adds type-specific metadata (season/episode for TV, artist for music, author for books).

### Source: `pkg/metadata/metadata.go`

`FormatFileSize` converts bytes to human-readable strings (KB, MB, GB, TB). `NormalizeTitle` lowercases, replaces separators with spaces, removes leading articles ("the", "a", "an"), and collapses whitespace. `SanitizeFilename` replaces OS-invalid characters with underscores.

### Source: `pkg/models/models.go`

`MediaItem` supports self-referential hierarchy via `ParentID` (TV show -> season -> episode, music artist -> album -> song). `QualityInfo` includes resolution, bitrate, codecs, frame rate, HDR, source, and a computed quality score.

## Practice Exercise

1. Analyze a set of filenames representing a TV season and verify that each result contains the correct season/episode numbers and show title.
2. Use `FormatFileSize` to format various file sizes. Use `NormalizeTitle` to compare titles with different separators and verify they match.
3. Populate a `models.MediaItem` hierarchy for a TV show with two seasons, each containing three episodes. Set parent IDs correctly.
