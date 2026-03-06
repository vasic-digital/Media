# Lesson 1: Media Type Detection with Rule-Based Engine

## Learning Objectives

- Build a rule-based detection engine with prioritized rules
- Detect TV shows, movies, music, books, photos, and software from filenames
- Add custom detection rules with match and extract functions

## Key Concepts

- **Rule Structure**: Each `Rule` has a `Name`, `Type`, `Priority`, `Match` function (returns bool), and `Extract` function (returns `*Detection`). The engine evaluates rules sorted by priority descending.
- **Confidence Scoring**: Each detection includes a confidence score (0.0 to 1.0). TV show patterns with both season and episode numbers score 0.95, while generic extension matches score 0.7-0.85.
- **Pattern Matching**: TV shows use regex patterns (`S01E02`, `1x02`, `Season 1 Episode 2`). Quality tags (resolution, source, codec, HDR) are extracted from release naming conventions.

## Code Walkthrough

### Source: `pkg/detector/detector.go`

The `Detect` method sorts rules by priority, runs each rule's `Match` function, and keeps the result with the highest confidence:

```go
func (e *Engine) Detect(filename string) *Detection {
    sorted := make([]Rule, len(e.rules))
    copy(sorted, e.rules)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Priority > sorted[j].Priority
    })
    // ...evaluate each rule, keep best confidence
}
```

The `extractTVShow` function parses season and episode numbers from regex submatches, extracts the title from the portion before the pattern match, and cleans release info from the title.

The `extractQualityTags` function identifies resolution (480p to 4K), source (BluRay, WEB-DL), video codec (H.264/H.265), audio codec (DTS, AAC, AC3), and HDR status from filename substrings.

## Practice Exercise

1. Create a detection engine and test it with filenames for each of the 7 media types. Verify the returned type, name, and confidence for each.
2. Add a custom rule for detecting audiobook files (`.m4b` extension) with priority 55. Verify it takes precedence over generic audio detection.
3. Test TV show detection with various patterns: `S01E02`, `1x03`, `Season.2.Episode.5`. Verify season and episode extraction for each format.
