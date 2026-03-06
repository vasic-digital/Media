# Getting Started

## Installation

```bash
go get digital.vasic.media
```

## Detecting Media Types

Use the detection engine to identify media type from a filename:

```go
package main

import (
    "fmt"

    "digital.vasic.media/pkg/detector"
)

func main() {
    engine := detector.NewEngine()

    // Detect a movie
    det := engine.Detect("The.Matrix.1999.1080p.BluRay.x264.mkv")
    fmt.Printf("Type: %s, Title: %q, Year: %d\n", det.Type, det.Name, det.Year)
    // Type: movie, Title: "The Matrix", Year: 1999

    // Detect a TV show
    det = engine.Detect("Breaking.Bad.S01E01.720p.mkv")
    fmt.Printf("Type: %s, Title: %q, S%02dE%02d\n",
        det.Type, det.Name, det.Season, det.Episode)
    // Type: tv_show, Title: "Breaking Bad", S01E01

    // Detect music
    det = engine.Detect("Pink Floyd - Comfortably Numb.flac")
    fmt.Printf("Type: %s, Title: %q, Artist: %s\n",
        det.Type, det.Name, det.Tags["artist"])
    // Type: music, Title: "Comfortably Numb", Artist: Pink Floyd
}
```

## Analyzing Files for Metadata

Extract structured metadata including resolution and codec:

```go
import "digital.vasic.media/pkg/analyzer"

a := analyzer.NewFilenameAnalyzer()
meta, err := a.Analyze("/media/Inception.2010.2160p.UHD.BluRay.x265.mkv")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Title: %s, Year: %d, Resolution: %s, Codec: %s\n",
    meta.Title, meta.Year, meta.Resolution, meta.Codec)
// Title: Inception, Year: 2010, Resolution: 2160p, Codec: H.265
```

## Using the Detection Pipeline

The pipeline combines detection, analysis, and provider enrichment:

```go
import (
    "context"
    "digital.vasic.media/pkg/manager"
    "digital.vasic.media/pkg/provider"
)

registry := provider.NewRegistry()
// registry.Register(myTMDBProvider)

pipeline := manager.NewPipeline(registry, nil)
result := pipeline.Process(context.Background(), "Interstellar.2014.1080p.mkv")

fmt.Printf("Detection: %s, Metadata title: %s\n",
    result.Detection.Type, result.Metadata.Title)
```
