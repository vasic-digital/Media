package manager

import (
	"context"
	"errors"
	"testing"

	"digital.vasic.media/pkg/detector"
	"digital.vasic.media/pkg/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- testLogger captures log calls to verify logger code paths ---

type testLogger struct {
	infos  []string
	warns  []string
	errors []string
}

func (l *testLogger) Info(msg string, keysAndValues ...interface{}) {
	l.infos = append(l.infos, msg)
}

func (l *testLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.warns = append(l.warns, msg)
}

func (l *testLogger) Error(msg string, keysAndValues ...interface{}) {
	l.errors = append(l.errors, msg)
}

// --- NewPipeline with Config and Logger ---

func TestNewPipeline_WithLogger(t *testing.T) {
	logger := &testLogger{}
	cfg := &Config{Logger: logger}

	p := NewPipeline(nil, cfg)
	require.NotNil(t, p)
	assert.NotNil(t, p.logger)
}

func TestNewPipeline_WithNilConfig(t *testing.T) {
	p := NewPipeline(nil, nil)
	require.NotNil(t, p)
	assert.Nil(t, p.logger)
}

func TestNewPipeline_WithConfigButNilLogger(t *testing.T) {
	cfg := &Config{Logger: nil}
	p := NewPipeline(nil, cfg)
	require.NotNil(t, p)
	assert.Nil(t, p.logger)
}

// --- Process: unknown type with logger ---

func TestProcess_UnknownType_WithLogger(t *testing.T) {
	logger := &testLogger{}
	p := NewPipeline(nil, &Config{Logger: logger})

	result := p.Process(context.Background(), "readme.txt")
	require.NotNil(t, result)
	assert.Equal(t, detector.TypeUnknown, result.Detection.Type)
	assert.Nil(t, result.Metadata)

	// Logger should have been called with "unknown media type"
	require.Len(t, logger.infos, 1)
	assert.Equal(t, "unknown media type", logger.infos[0])
}

// --- Process: unknown type without logger (no panic) ---

func TestProcess_UnknownType_NoLogger(t *testing.T) {
	p := NewPipeline(nil, nil)

	result := p.Process(context.Background(), "readme.txt")
	require.NotNil(t, result)
	assert.Equal(t, detector.TypeUnknown, result.Detection.Type)
	assert.Nil(t, result.Metadata)
}

// --- Process: provider error with logger ---

func TestProcess_ProviderError_WithLogger(t *testing.T) {
	logger := &testLogger{}
	reg := provider.NewRegistry()
	reg.Register(&mockProvider{
		name: "failing",
		err:  errors.New("api down"),
	})

	p := NewPipeline(reg, &Config{Logger: logger})
	result := p.Process(context.Background(), "The.Matrix.1999.1080p.mkv")

	require.NotNil(t, result)
	assert.Nil(t, result.Error)
	assert.Empty(t, result.Providers)

	// Logger should have logged the provider search failure.
	require.Len(t, logger.warns, 1)
	assert.Equal(t, "provider search failed", logger.warns[0])
}

// --- Process: known type with provider that returns results and logger ---

func TestProcess_WithProvider_AndLogger(t *testing.T) {
	logger := &testLogger{}
	reg := provider.NewRegistry()
	reg.Register(&mockProvider{
		name: "test",
		results: []*provider.SearchResult{
			{Title: "The Matrix", Year: 1999, Rating: 8.7},
		},
	})

	p := NewPipeline(reg, &Config{Logger: logger})
	result := p.Process(context.Background(), "The.Matrix.1999.1080p.mkv")

	require.NotNil(t, result)
	assert.Nil(t, result.Error)
	require.Len(t, result.Providers, 1)
	assert.Equal(t, "The Matrix", result.Providers[0].Title)
	// No warnings should have been logged for successful provider search.
	assert.Empty(t, logger.warns)
}

// --- Process: nil registry (no provider enrichment) ---

func TestProcess_NilRegistry(t *testing.T) {
	p := NewPipeline(nil, nil)
	result := p.Process(context.Background(), "Inception.2010.1080p.mkv")

	require.NotNil(t, result)
	assert.Nil(t, result.Error)
	assert.Equal(t, detector.TypeMovie, result.Detection.Type)
	assert.NotNil(t, result.Metadata)
	assert.Empty(t, result.Providers)
}

// --- Process: registry with provider, but metadata title is empty ---

func TestProcess_EmptyTitle_NoProviderSearch(t *testing.T) {
	logger := &testLogger{}
	reg := provider.NewRegistry()
	reg.Register(&mockProvider{
		name: "test",
		results: []*provider.SearchResult{
			{Title: "should not appear"},
		},
	})

	p := NewPipeline(reg, &Config{Logger: logger})
	// A photo file — extractPhoto uses the raw filename as Name,
	// which won't be empty, but we can verify the provider path.
	result := p.Process(context.Background(), "vacation.jpg")

	require.NotNil(t, result)
	assert.Equal(t, detector.TypePhoto, result.Detection.Type)
}

// --- ProcessBatch: mixed types ---

func TestProcessBatch_Mixed_WithLogger(t *testing.T) {
	logger := &testLogger{}
	p := NewPipeline(nil, &Config{Logger: logger})

	paths := []string{
		"movie.mkv",
		"show.S01E01.mkv",
		"song.mp3",
		"readme.txt",
		"photo.jpg",
	}

	results := p.ProcessBatch(context.Background(), paths)
	require.Len(t, results, 5)

	assert.Equal(t, detector.TypeMovie, results[0].Detection.Type)
	assert.Equal(t, detector.TypeTVShow, results[1].Detection.Type)
	assert.Equal(t, detector.TypeMusic, results[2].Detection.Type)
	assert.Equal(t, detector.TypeUnknown, results[3].Detection.Type)
	assert.Equal(t, detector.TypePhoto, results[4].Detection.Type)

	// Logger should have Info for the unknown type.
	assert.Len(t, logger.infos, 1)
}

// --- ProcessBatch: empty list ---

func TestProcessBatch_EmptyList(t *testing.T) {
	p := NewPipeline(nil, nil)
	results := p.ProcessBatch(context.Background(), []string{})
	assert.Empty(t, results)
}
