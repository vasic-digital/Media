package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements the Provider interface for testing.
type mockProvider struct {
	name    string
	results []*SearchResult
	err     error
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Search(ctx context.Context, query string) ([]*SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Filter by query for realism
	var filtered []*SearchResult
	for _, r := range m.results {
		filtered = append(filtered, r)
	}
	return filtered, nil
}

func (m *mockProvider) GetByID(ctx context.Context, id string) (*SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, r := range m.results {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, fmt.Errorf("not found: %s", id)
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	require.NotNil(t, r)
	assert.Empty(t, r.List())
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	p := &mockProvider{name: "test_provider"}
	r.Register(p)

	names := r.List()
	assert.Contains(t, names, "test_provider")
}

func TestRegistry_RegisterOverwrite(t *testing.T) {
	r := NewRegistry()

	p1 := &mockProvider{
		name: "tmdb",
		results: []*SearchResult{
			{ID: "1", Title: "Old"},
		},
	}
	p2 := &mockProvider{
		name: "tmdb",
		results: []*SearchResult{
			{ID: "2", Title: "New"},
		},
	}

	r.Register(p1)
	r.Register(p2)

	names := r.List()
	assert.Len(t, names, 1)

	got, ok := r.Get("tmdb")
	require.True(t, ok)

	results, err := got.Search(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, "New", results[0].Title)
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()
	p := &mockProvider{name: "tmdb"}
	r.Register(p)

	got, ok := r.Get("tmdb")
	assert.True(t, ok)
	assert.Equal(t, "tmdb", got.Name())

	_, ok = r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockProvider{name: "tmdb"})
	r.Register(&mockProvider{name: "imdb"})
	r.Register(&mockProvider{name: "musicbrainz"})

	names := r.List()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "tmdb")
	assert.Contains(t, names, "imdb")
	assert.Contains(t, names, "musicbrainz")
}

func TestRegistry_Search(t *testing.T) {
	r := NewRegistry()

	r.Register(&mockProvider{
		name: "tmdb",
		results: []*SearchResult{
			{ID: "t1", Title: "The Matrix", Year: 1999, Rating: 8.7},
			{ID: "t2", Title: "Matrix Reloaded", Year: 2003, Rating: 7.2},
		},
	})

	r.Register(&mockProvider{
		name: "imdb",
		results: []*SearchResult{
			{ID: "i1", Title: "The Matrix", Year: 1999, Rating: 8.7},
		},
	})

	results, err := r.Search(context.Background(), "matrix")
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Verify source tagging
	sourceCount := map[string]int{}
	for _, sr := range results {
		sourceCount[sr.Source]++
	}
	assert.Equal(t, 2, sourceCount["tmdb"])
	assert.Equal(t, 1, sourceCount["imdb"])
}

func TestRegistry_SearchEmpty(t *testing.T) {
	r := NewRegistry()

	results, err := r.Search(context.Background(), "anything")
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestRegistry_SearchWithErrors(t *testing.T) {
	r := NewRegistry()

	// One provider fails, one succeeds
	r.Register(&mockProvider{
		name: "failing",
		err:  fmt.Errorf("api timeout"),
	})

	r.Register(&mockProvider{
		name: "working",
		results: []*SearchResult{
			{ID: "w1", Title: "Result"},
		},
	})

	results, err := r.Search(context.Background(), "test")
	require.NoError(t, err) // Should not fail because one provider succeeded
	assert.Len(t, results, 1)
	assert.Equal(t, "Result", results[0].Title)
}

func TestRegistry_SearchAllFail(t *testing.T) {
	r := NewRegistry()

	r.Register(&mockProvider{
		name: "p1",
		err:  fmt.Errorf("error 1"),
	})

	r.Register(&mockProvider{
		name: "p2",
		err:  fmt.Errorf("error 2"),
	})

	results, err := r.Search(context.Background(), "test")
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "all providers failed")
}

func TestRegistry_SearchContextCancelled(t *testing.T) {
	r := NewRegistry()

	r.Register(&mockProvider{
		name: "slow",
		results: []*SearchResult{
			{ID: "1", Title: "Result"},
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The mock provider does not check context, so it will still return results.
	// This test verifies the function handles a cancelled context gracefully.
	results, err := r.Search(ctx, "test")
	// May or may not error depending on timing, but should not panic
	_ = results
	_ = err
}

func TestProvider_GetByID(t *testing.T) {
	p := &mockProvider{
		name: "tmdb",
		results: []*SearchResult{
			{ID: "123", Title: "The Matrix", Year: 1999},
			{ID: "456", Title: "Inception", Year: 2010},
		},
	}

	result, err := p.GetByID(context.Background(), "123")
	require.NoError(t, err)
	assert.Equal(t, "The Matrix", result.Title)
	assert.Equal(t, 1999, result.Year)

	_, err = p.GetByID(context.Background(), "999")
	assert.Error(t, err)
}

func TestProvider_GetByIDError(t *testing.T) {
	p := &mockProvider{
		name: "failing",
		err:  fmt.Errorf("connection refused"),
	}

	_, err := p.GetByID(context.Background(), "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestSearchResult_Fields(t *testing.T) {
	sr := &SearchResult{
		ID:          "tt0133093",
		Title:       "The Matrix",
		Year:        1999,
		Description: "A computer hacker learns about the true nature of reality.",
		PosterURL:   "https://example.com/matrix.jpg",
		Rating:      8.7,
		Source:      "imdb",
	}

	assert.Equal(t, "tt0133093", sr.ID)
	assert.Equal(t, "The Matrix", sr.Title)
	assert.Equal(t, 1999, sr.Year)
	assert.Equal(t, 8.7, sr.Rating)
	assert.Equal(t, "imdb", sr.Source)
	assert.NotEmpty(t, sr.Description)
	assert.NotEmpty(t, sr.PosterURL)
}
