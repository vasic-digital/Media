# FAQ

## What media types does the detector recognize?

The detector identifies 7 media types: `movie`, `tv_show`, `music`, `book`, `photo`, `game`, and `software`. TV show detection uses patterns like `S01E02`, `1x02`, and `Season 1 Episode 2`. Other types are detected by file extension (e.g., `.mkv` for movies, `.flac` for music, `.epub` for books).

## How does detection priority work?

Rules have a priority field (integer). Higher priority rules are evaluated first. TV show patterns have priority 100 because they override generic video extension matching (a `.mkv` file with `S01E02` in the name is a TV show, not a movie). Extension-based rules default to priority 50. Custom rules can use any priority.

## Is the provider registry thread-safe?

Yes. The `provider.Registry` uses `sync.RWMutex` for all operations. `Register()` acquires a write lock, while `Get()`, `List()`, and `Search()` use read locks. The `Search()` method queries all providers concurrently using goroutines and aggregates results.

## What happens when a provider fails during search?

The registry's `Search()` method uses graceful error handling. If some providers fail but others succeed, the successful results are returned without error. An error is only returned if all registered providers fail. Individual failures are collected and included in the aggregate error message.

## Can I use the detection engine without the full pipeline?

Yes. Use `detector.NewEngine()` and call `engine.Detect(filename)` directly for detection only. Use `analyzer.NewFilenameAnalyzer()` for metadata extraction without provider enrichment. The `manager.Pipeline` is a convenience facade that combines all three steps, but each package works independently.
