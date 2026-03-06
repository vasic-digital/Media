# Lesson 4: Pipeline Facade and Domain Models

## Learning Objectives

- Use the Facade pattern to hide the three-step detect-analyze-enrich workflow
- Understand the domain model structs for the media entity system
- Process files in batch with the pipeline

## Key Concepts

- **Facade Pattern**: `Pipeline` wires detector, analyzer, and provider registry into a single `Process(ctx, path)` call. The caller does not need to know about the three internal stages.
- **Domain Models**: The `models` package defines `MediaItem`, `MediaFile`, `MediaCollection`, `SearchRequest`, and related structs with JSON serialization. These represent the full media entity graph.
- **Batch Processing**: `ProcessBatch(ctx, paths)` processes multiple files, returning a result for each with detection, metadata, and provider results.

## Code Walkthrough

### Source: `pkg/manager/manager.go`

The pipeline orchestrates:

```
Path -> [1] detector.Engine.Detect() -> Detection
     -> [2] analyzer.Analyze()       -> Metadata
     -> [3] provider.Registry.Search() -> []*SearchResult
     -> Result{Path, Detection, Metadata, Providers, Error}
```

The `NewPipeline` constructor wires the components together. A `Logger` interface allows injecting any structured logger.

### Source: `pkg/models/models.go`

Domain model structs:
- `MediaItem` -- central entity with type, title, year, parent_id for hierarchy
- `MediaFile` -- junction to scanned files
- `MediaCollection` -- user-created groupings
- `SearchRequest` -- parameters for provider queries
- `QualityInfo` -- resolution, codec, bitrate

### Source: `pkg/manager/manager_test.go` and `pkg/models/models_test.go`

Tests verify pipeline end-to-end flow, batch processing, and model serialization.

## Practice Exercise

1. Create a `Pipeline` with a detection engine (default rules) and a mock provider registry. Process the file `/media/The.Godfather.1972.1080p.BluRay.mkv`. Verify the result contains type=movie, title="The Godfather", year=1972.
2. Use `ProcessBatch` with 5 different files (movie, TV show, music, book, software). Verify each result has the correct media type and metadata.
3. Create `MediaItem` and `MediaFile` model instances and serialize them to JSON. Verify the JSON structure matches the expected schema with all fields populated.
