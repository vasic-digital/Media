# Lesson 3: External Providers and the Detection Pipeline

## Learning Objectives

- Implement the Provider interface for external metadata sources
- Use the Registry for thread-safe provider management and concurrent search
- Orchestrate detection, analysis, and enrichment with the Pipeline

## Key Concepts

- **Provider Interface**: Providers implement `Name() string`, `Search(ctx, query) ([]*SearchResult, error)`, and `GetByID(ctx, id) (*SearchResult, error)`. The registry manages registration and lookup.
- **Concurrent Search**: `Registry.Search()` queries all providers concurrently using goroutines and a results channel. Partial failures do not block -- results from successful providers are returned even if others fail.
- **Pipeline Facade**: `manager.Pipeline` orchestrates three steps: (1) detect media type, (2) analyze filename for metadata, (3) search external providers with the extracted title. `ProcessBatch` handles multiple files with context cancellation.

## Code Walkthrough

### Source: `pkg/provider/provider.go`

The `Search` method launches one goroutine per provider, collects results through a channel, and tags each result with its source provider name:

```go
ch := make(chan result, len(providers))
for _, p := range providers {
    go func(p Provider) {
        results, err := p.Search(ctx, query)
        ch <- result{results: results, err: err, name: p.Name()}
    }(p)
}
```

### Source: `pkg/manager/manager.go`

The `Process` method runs the three-step pipeline. If detection returns `TypeUnknown`, it returns early. If analysis fails, it sets the error but still returns the detection result. Provider search is optional (skipped if no registry or empty title).

`ProcessBatch` iterates paths, checking for context cancellation between each file.

## Practice Exercise

1. Implement a mock `Provider` that returns hardcoded search results. Register it with a `Registry` and call `Search()`. Verify the results include the provider's source name.
2. Create a pipeline with the mock provider. Process a movie filename and verify that `result.Providers` contains enrichment data from the mock.
3. Use `ProcessBatch` with a context that cancels after 3 files. Verify that remaining files in the batch have `ctx.Err()` set.
