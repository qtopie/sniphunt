package search

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
)

func (s *Searcher) walk(ctx context.Context, root string, tasks chan<- task) error {
	return godirwalk.Walk(root, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				if s.shouldIgnore(osPathname, de.Name()) {
					return filepath.SkipDir
				}
				return nil
			}

			if s.shouldProcess(osPathname, de.Name()) {
				select {
				case tasks <- task{path: osPathname}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		},
		Unsorted: true, // Performance boost
	})
}

func (s *Searcher) shouldIgnore(path, name string) bool {
	if !s.IncludeHidden && strings.HasPrefix(name, ".") && name != "." {
		return true
	}
	for _, ignore := range s.Ignore {
		if strings.Contains(path, ignore) {
			return true
		}
	}
	return false
}

func (s *Searcher) shouldProcess(path, name string) bool {
	if s.shouldIgnore(path, name) {
		return false
	}
	if len(s.Extensions) == 0 {
		return true
	}
	ext := filepath.Ext(name)
	for _, e := range s.Extensions {
		if ext == e {
			return true
		}
	}
	return false
}
