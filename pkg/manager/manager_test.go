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

func TestNewPipeline(t *testing.T) {
	p := NewPipeline(nil, nil)
	assert.NotNil(t, p)
	assert.NotNil(t, p.engine)
	assert.NotNil(t, p.analyzer)
}

func TestPipeline_Process_Movie(t *testing.T) {
	p := NewPipeline(nil, nil)
	result := p.Process(context.Background(), "The.Matrix.1999.1080p.BluRay.mkv")

	require.NotNil(t, result)
	assert.Nil(t, result.Error)
	assert.Equal(t, detector.TypeMovie, result.Detection.Type)
	assert.NotNil(t, result.Metadata)
	assert.Contains(t, result.Metadata.Title, "Matrix")
}

func TestPipeline_Process_TVShow(t *testing.T) {
	p := NewPipeline(nil, nil)
	result := p.Process(context.Background(), "Breaking.Bad.S01E01.720p.mkv")

	require.NotNil(t, result)
	assert.Nil(t, result.Error)
	assert.Equal(t, detector.TypeTVShow, result.Detection.Type)
	assert.NotNil(t, result.Metadata)
}

func TestPipeline_Process_Music(t *testing.T) {
	p := NewPipeline(nil, nil)
	result := p.Process(context.Background(), "Pink Floyd - Comfortably Numb.flac")

	require.NotNil(t, result)
	assert.Nil(t, result.Error)
	assert.Equal(t, detector.TypeMusic, result.Detection.Type)
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, "Comfortably Numb", result.Metadata.Title)
}

func TestPipeline_Process_Unknown(t *testing.T) {
	p := NewPipeline(nil, nil)
	result := p.Process(context.Background(), "readme.txt")

	require.NotNil(t, result)
	assert.Equal(t, detector.TypeUnknown, result.Detection.Type)
	assert.Nil(t, result.Metadata) // No analysis for unknown type
}

type mockProvider struct {
	name    string
	results []*provider.SearchResult
	err     error
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Search(ctx context.Context, query string) ([]*provider.SearchResult, error) {
	return m.results, m.err
}
func (m *mockProvider) GetByID(ctx context.Context, id string) (*provider.SearchResult, error) {
	return nil, nil
}

func TestPipeline_Process_WithProvider(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register(&mockProvider{
		name: "test",
		results: []*provider.SearchResult{
			{Title: "The Matrix", Year: 1999, Rating: 8.7},
		},
	})

	p := NewPipeline(reg, nil)
	result := p.Process(context.Background(), "The.Matrix.1999.1080p.mkv")

	require.NotNil(t, result)
	assert.Nil(t, result.Error)
	require.Len(t, result.Providers, 1)
	assert.Equal(t, "The Matrix", result.Providers[0].Title)
}

func TestPipeline_Process_ProviderError(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register(&mockProvider{
		name: "failing",
		err:  errors.New("api down"),
	})

	p := NewPipeline(reg, nil)
	result := p.Process(context.Background(), "The.Matrix.1999.1080p.mkv")

	require.NotNil(t, result)
	// Provider errors don't fail the pipeline
	assert.Nil(t, result.Error)
	assert.Empty(t, result.Providers)
}

func TestPipeline_ProcessBatch(t *testing.T) {
	p := NewPipeline(nil, nil)
	paths := []string{
		"The.Matrix.1999.mkv",
		"Breaking.Bad.S01E01.mkv",
		"readme.txt",
	}

	results := p.ProcessBatch(context.Background(), paths)
	assert.Len(t, results, 3)
	assert.Equal(t, detector.TypeMovie, results[0].Detection.Type)
	assert.Equal(t, detector.TypeTVShow, results[1].Detection.Type)
	assert.Equal(t, detector.TypeUnknown, results[2].Detection.Type)
}

func TestPipeline_ProcessBatch_ContextCancelled(t *testing.T) {
	p := NewPipeline(nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	paths := []string{"file1.mkv", "file2.mkv"}
	results := p.ProcessBatch(ctx, paths)
	assert.Len(t, results, 2)

	// At least one should have context error
	hasContextErr := false
	for _, r := range results {
		if r.Error == context.Canceled {
			hasContextErr = true
			break
		}
	}
	assert.True(t, hasContextErr)
}

func TestPipeline_DetectOnly(t *testing.T) {
	p := NewPipeline(nil, nil)
	det := p.DetectOnly("movie.mkv")
	assert.Equal(t, detector.TypeMovie, det.Type)
}

func TestPipeline_Engine(t *testing.T) {
	p := NewPipeline(nil, nil)
	assert.NotNil(t, p.Engine())
}
