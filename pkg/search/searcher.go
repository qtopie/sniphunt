package search

import (
	"context"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"
)

// Match represents a single search result.
type Match struct {
	Path    string
	LineNum int
	Text    []byte
	Score   float64 // Used for similarity search
}

// Searcher holds the configuration for a search operation.
type Searcher struct {
	Workers       int
	Ignore        []string
	Extensions    []string
	IncludeHidden bool
}

// NewSearcher creates a new Searcher with default values.
func NewSearcher() *Searcher {
	return &Searcher{
		Workers: runtime.NumCPU(),
	}
}

// task represents a file to be processed by a worker.
type task struct {
	path string
}

// Search executes a regex search starting from the root directory.
func (s *Searcher) Search(ctx context.Context, root string, patternStr string) (<-chan Match, <-chan error) {
	matchChan := make(chan Match, 1000)
	errChan := make(chan error, 1)

	re, err := regexp.Compile(patternStr)
	if err != nil {
		errChan <- err
		close(matchChan)
		close(errChan)
		return matchChan, errChan
	}

	// Simple literal extraction: if no special chars, use as literal filter
	literal := ""
	if !strings.ContainsAny(patternStr, ".*+?()|[]{}^$\\") {
		literal = patternStr
	}

	go func() {
		defer close(matchChan)
		defer close(errChan)

		g, ctx := errgroup.WithContext(ctx)
		tasks := make(chan task, s.Workers*10)

		// Start producer
		g.Go(func() error {
			defer close(tasks)
			return s.walk(ctx, root, tasks)
		})

		// Start workers
		for i := 0; i < s.Workers; i++ {
			g.Go(func() error {
				return s.regexWorker(ctx, re, literal, tasks, matchChan)
			})
		}

		if err := g.Wait(); err != nil {
			errChan <- err
		}
	}()

	return matchChan, errChan
}

// MatchSnippet executes a similarity-based search starting from the root directory.
func (s *Searcher) MatchSnippet(ctx context.Context, root string, snippet string) (<-chan Match, <-chan error) {
	matchChan := make(chan Match, 1000)
	errChan := make(chan error, 1)

	go func() {
		defer close(matchChan)
		defer close(errChan)

		g, ctx := errgroup.WithContext(ctx)
		tasks := make(chan task, s.Workers*10)

		// Start producer
		g.Go(func() error {
			defer close(tasks)
			return s.walk(ctx, root, tasks)
		})

		// Start workers
		for i := 0; i < s.Workers; i++ {
			g.Go(func() error {
				return s.similarityWorker(ctx, snippet, tasks, matchChan)
			})
		}

		if err := g.Wait(); err != nil {
			errChan <- err
		}
	}()

	return matchChan, errChan
}
