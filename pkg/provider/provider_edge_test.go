package provider

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// copyingMockProvider returns fresh copies of results on each Search call
// to avoid data races when Registry.Search sets sr.Source concurrently.
type copyingMockProvider struct {
	name    string
	results []*SearchResult
	err     error
}

func (m *copyingMockProvider) Name() string { return m.name }

func (m *copyingMockProvider) Search(ctx context.Context, query string) ([]*SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	copies := make([]*SearchResult, len(m.results))
	for i, r := range m.results {
		cp := *r
		copies[i] = &cp
	}
	return copies, nil
}

func (m *copyingMockProvider) GetByID(ctx context.Context, id string) (*SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, r := range m.results {
		if r.ID == id {
			cp := *r
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("not found: %s", id)
}

// --- Concurrent Register and Search ---

func TestRegistry_ConcurrentRegisterAndSearch(t *testing.T) {
	r := NewRegistry()

	const numProviders = 20
	const numSearchers = 30
	var wg sync.WaitGroup
	wg.Add(numProviders + numSearchers)

	// Concurrently register providers
	for i := 0; i < numProviders; i++ {
		go func(idx int) {
			defer wg.Done()
			r.Register(&copyingMockProvider{
				name: fmt.Sprintf("provider_%d", idx),
				results: []*SearchResult{
					{ID: fmt.Sprintf("id_%d", idx), Title: fmt.Sprintf("Result %d", idx)},
				},
			})
		}(i)
	}

	// Concurrently search the registry
	searchErrors := make(chan error, numSearchers)
	for i := 0; i < numSearchers; i++ {
		go func() {
			defer wg.Done()
			_, err := r.Search(context.Background(), "test query")
			// err is acceptable if all providers fail, but should not panic
			if err != nil {
				searchErrors <- err
			}
		}()
	}

	wg.Wait()
	close(searchErrors)

	// Registry should contain all registered providers (some may have
	// overwritten if names collide, but all unique names here)
	names := r.List()
	assert.True(t, len(names) > 0, "at least some providers should be registered")
	assert.LessOrEqual(t, len(names), numProviders)
}

// --- Concurrent Register, Get, List, and Search ---

func TestRegistry_ConcurrentAllOperations(t *testing.T) {
	r := NewRegistry()

	// Pre-register a few providers (use copyingMockProvider to avoid races)
	for i := 0; i < 5; i++ {
		r.Register(&copyingMockProvider{
			name: fmt.Sprintf("initial_%d", i),
			results: []*SearchResult{
				{ID: fmt.Sprintf("init_%d", i), Title: fmt.Sprintf("Initial %d", i)},
			},
		})
	}

	const ops = 100
	var wg sync.WaitGroup
	wg.Add(ops)

	for i := 0; i < ops; i++ {
		go func(idx int) {
			defer wg.Done()
			switch idx % 4 {
			case 0: // Register
				r.Register(&copyingMockProvider{
					name: fmt.Sprintf("dynamic_%d", idx),
					results: []*SearchResult{
						{ID: fmt.Sprintf("dyn_%d", idx), Title: "Dynamic"},
					},
				})
			case 1: // Get
				_, _ = r.Get(fmt.Sprintf("initial_%d", idx%5))
			case 2: // List
				_ = r.List()
			case 3: // Search
				_, _ = r.Search(context.Background(), "query")
			}
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

// --- Duplicate Registration ---

func TestRegistry_DuplicateRegistration(t *testing.T) {
	r := NewRegistry()

	original := &mockProvider{
		name: "tmdb",
		results: []*SearchResult{
			{ID: "1", Title: "Original Result", Year: 2000},
		},
	}
	replacement := &mockProvider{
		name: "tmdb",
		results: []*SearchResult{
			{ID: "2", Title: "Replacement Result", Year: 2020},
		},
	}

	// Register original
	r.Register(original)
	names := r.List()
	assert.Len(t, names, 1)

	// Register replacement with same name
	r.Register(replacement)
	names = r.List()
	assert.Len(t, names, 1, "duplicate name should overwrite, not add")

	// Verify the replacement is active
	p, ok := r.Get("tmdb")
	require.True(t, ok)
	results, err := p.Search(context.Background(), "test")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Replacement Result", results[0].Title)
	assert.Equal(t, 2020, results[0].Year)
}

// --- Search No Match ---

func TestRegistry_SearchNoMatch(t *testing.T) {
	r := NewRegistry()

	// Register a provider that returns empty results
	r.Register(&mockProvider{
		name:    "empty_provider",
		results: []*SearchResult{}, // no results
	})

	results, err := r.Search(context.Background(), "nonexistent query xyz123")
	require.NoError(t, err)
	// The mock provider returns empty slice which is appended to allResults
	assert.Empty(t, results)
}

// --- Empty Registry ---

func TestRegistry_EmptyRegistry(t *testing.T) {
	r := NewRegistry()

	t.Run("list_empty", func(t *testing.T) {
		names := r.List()
		assert.Empty(t, names)
		assert.NotNil(t, names) // should be empty slice, not nil
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		p, ok := r.Get("anything")
		assert.False(t, ok)
		assert.Nil(t, p)
	})

	t.Run("search_empty_registry", func(t *testing.T) {
		results, err := r.Search(context.Background(), "test")
		assert.NoError(t, err)
		assert.Nil(t, results) // empty registry returns nil, nil
	})

	t.Run("search_empty_query_empty_registry", func(t *testing.T) {
		results, err := r.Search(context.Background(), "")
		assert.NoError(t, err)
		assert.Nil(t, results)
	})
}

// --- Search with empty query string ---

func TestRegistry_SearchEmptyQuery(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockProvider{
		name: "test",
		results: []*SearchResult{
			{ID: "1", Title: "Some Result"},
		},
	})

	// Empty query should still call providers (mock returns all results)
	results, err := r.Search(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

// --- Multiple providers, some returning nil results ---

func TestRegistry_SearchProviderReturnsNilResults(t *testing.T) {
	r := NewRegistry()

	r.Register(&mockProvider{
		name:    "nil_results",
		results: nil, // nil, not empty slice
	})
	r.Register(&mockProvider{
		name: "valid_results",
		results: []*SearchResult{
			{ID: "1", Title: "Valid"},
		},
	})

	results, err := r.Search(context.Background(), "test")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Valid", results[0].Title)
}
