package sense

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

var ErrNotFound = errors.New("no matching elements found")

// Finder tries sources in priority order and returns the first hit.
type Finder struct {
	sources []Source
}

// NewFinder creates a Finder with the given sources, sorted by priority (lowest first).
func NewFinder(sources ...Source) *Finder {
	sorted := make([]Source, len(sources))
	copy(sorted, sources)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority() < sorted[j].Priority()
	})
	return &Finder{sources: sorted}
}

// Find runs the query against sources in priority order.
// Stops at the first source that returns results.
// Skips sources that report unavailable.
// Returns ErrNotFound if no source returns results.
func (f *Finder) Find(ctx context.Context, q Query) ([]Match, error) {
	for _, src := range f.sources {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if !src.Available(ctx) {
			continue
		}
		matches, err := src.Find(ctx, q)
		if err != nil {
			// Log but continue to next source
			continue
		}
		if len(matches) > 0 {
			if q.MaxResults > 0 && len(matches) > q.MaxResults {
				matches = matches[:q.MaxResults]
			}
			return matches, nil
		}
	}
	return nil, fmt.Errorf("%w: query=%q role=%s app=%s", ErrNotFound, q.Name, q.Role, q.AppID)
}

// FindAll runs the query against ALL available sources and merges results.
// Does not deduplicate — returns all matches from all sources.
func (f *Finder) FindAll(ctx context.Context, q Query) ([]Match, error) {
	var all []Match
	for _, src := range f.sources {
		if ctx.Err() != nil {
			return all, ctx.Err()
		}
		if !src.Available(ctx) {
			continue
		}
		matches, err := src.Find(ctx, q)
		if err != nil {
			continue
		}
		all = append(all, matches...)
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("%w: query=%q (all sources)", ErrNotFound, q.Name)
	}
	if q.MaxResults > 0 && len(all) > q.MaxResults {
		all = all[:q.MaxResults]
	}
	return all, nil
}

// Sources returns the registered sources in priority order.
func (f *Finder) Sources() []Source {
	return f.sources
}
