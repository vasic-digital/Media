// Package manager provides a media detection pipeline that orchestrates
// detection, analysis, and external metadata enrichment.
//
// It combines the detector, analyzer, and provider packages into a
// single pipeline that processes file paths and returns enriched
// media items.
//
// Design pattern: Facade (simplifies the multi-step pipeline).
package manager

import (
	"context"
	"fmt"

	"digital.vasic.media/pkg/analyzer"
	"digital.vasic.media/pkg/detector"
	"digital.vasic.media/pkg/provider"
)

// Logger is a minimal logging interface.
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// Result represents the output of the detection pipeline for a single file.
type Result struct {
	Path      string
	Detection *detector.Detection
	Metadata  *analyzer.Metadata
	Providers []*provider.SearchResult
	Error     error
}

// Pipeline orchestrates the media detection, analysis, and enrichment workflow.
type Pipeline struct {
	engine   *detector.Engine
	analyzer *analyzer.FilenameAnalyzer
	registry *provider.Registry
	logger   Logger
}

// Config holds pipeline configuration.
type Config struct {
	Logger Logger
}

// NewPipeline creates a new media detection pipeline.
func NewPipeline(registry *provider.Registry, cfg *Config) *Pipeline {
	p := &Pipeline{
		engine:   detector.NewEngine(),
		analyzer: analyzer.NewFilenameAnalyzer(),
		registry: registry,
	}
	if cfg != nil {
		p.logger = cfg.Logger
	}
	return p
}

// Process runs the full detection pipeline on a file path.
// It detects the media type, extracts metadata from the filename,
// and optionally queries external providers for enrichment.
func (p *Pipeline) Process(ctx context.Context, path string) *Result {
	result := &Result{Path: path}

	// Step 1: Detect media type
	result.Detection = p.engine.Detect(path)
	if result.Detection.Type == detector.TypeUnknown {
		if p.logger != nil {
			p.logger.Info("unknown media type", "path", path)
		}
		return result
	}

	// Step 2: Analyze filename for metadata
	meta, err := p.analyzer.Analyze(path)
	if err != nil {
		result.Error = fmt.Errorf("analysis failed: %w", err)
		if p.logger != nil {
			p.logger.Warn("analysis failed", "path", path, "error", err)
		}
		return result
	}
	result.Metadata = meta

	// Step 3: Enrich from external providers (if available and title is non-empty)
	if p.registry != nil && meta.Title != "" {
		results, err := p.registry.Search(ctx, meta.Title)
		if err != nil {
			if p.logger != nil {
				p.logger.Warn("provider search failed", "title", meta.Title, "error", err)
			}
		} else {
			result.Providers = results
		}
	}

	return result
}

// ProcessBatch runs the pipeline on multiple paths and returns all results.
func (p *Pipeline) ProcessBatch(ctx context.Context, paths []string) []*Result {
	results := make([]*Result, len(paths))
	for i, path := range paths {
		select {
		case <-ctx.Done():
			results[i] = &Result{Path: path, Error: ctx.Err()}
		default:
			results[i] = p.Process(ctx, path)
		}
	}
	return results
}

// DetectOnly runs only the detection step (no analysis or provider enrichment).
func (p *Pipeline) DetectOnly(path string) *detector.Detection {
	return p.engine.Detect(path)
}

// Engine returns the underlying detection engine for custom rule addition.
func (p *Pipeline) Engine() *detector.Engine {
	return p.engine
}
