// Package provider defines interfaces for external metadata providers
// such as TMDB, IMDB, MusicBrainz, etc.
package provider

import (
	"context"
	"fmt"
	"sync"
)

// SearchResult represents a metadata search result.
type SearchResult struct {
	ID          string
	Title       string
	Year        int
	Description string
	PosterURL   string
	Rating      float64
	Source      string
}

// Provider defines the interface for metadata providers.
type Provider interface {
	Name() string
	Search(ctx context.Context, query string) ([]*SearchResult, error)
	GetByID(ctx context.Context, id string) (*SearchResult, error)
}

// Registry holds registered metadata providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry.
// If a provider with the same name already exists, it is replaced.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// List returns the names of all registered providers.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Search queries all registered providers and aggregates results.
// Results are returned as a map keyed by provider name.
// Errors from individual providers are collected but do not prevent
// other providers from being queried. If all providers fail, an
// aggregated error is returned.
func (r *Registry) Search(ctx context.Context, query string) ([]*SearchResult, error) {
	r.mu.RLock()
	providers := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	r.mu.RUnlock()

	if len(providers) == 0 {
		return nil, nil
	}

	type result struct {
		results []*SearchResult
		err     error
		name    string
	}

	ch := make(chan result, len(providers))
	for _, p := range providers {
		go func(p Provider) {
			results, err := p.Search(ctx, query)
			ch <- result{results: results, err: err, name: p.Name()}
		}(p)
	}

	var allResults []*SearchResult
	var errors []error

	for range providers {
		res := <-ch
		if res.err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", res.name, res.err))
			continue
		}
		// Tag each result with its source provider
		for _, sr := range res.results {
			sr.Source = res.name
		}
		allResults = append(allResults, res.results...)
	}

	if len(allResults) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all providers failed: %v", errors)
	}

	return allResults, nil
}
