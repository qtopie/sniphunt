package search

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSearch(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sniphunt-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files
	files := map[string]string{
		"test1.java": "public class Test1 {\n  public void hello() {\n    System.out.println(\"hello world\");\n  }\n}",
		"test2.java": "package com.example;\n\nfunc main() {\n  println(\"another test\");\n}",
		"ignore.txt": "this should be ignored",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	s := NewSearcher()
	s.Extensions = []string{".java"}

	t.Run("Regex Search", func(t *testing.T) {
		ctx := context.Background()
		matchChan, errChan := s.Search(ctx, tmpDir, "hello")

		count := 0
		for {
			select {
			case _, ok := <-matchChan:
				if !ok {
					matchChan = nil
				} else {
					count++
				}
			case err, ok := <-errChan:
				if !ok {
					errChan = nil
				} else if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
			if matchChan == nil && errChan == nil {
				break
			}
		}

		if count != 2 {
			t.Errorf("Expected 2 matches, got %d", count)
		}
	})

	t.Run("Similarity Search", func(t *testing.T) {
		ctx := context.Background()
		target := "public void hello() { System.out.println(\"hello world\"); }"
		matchChan, errChan := s.MatchSnippet(ctx, tmpDir, target)

		var results []Match
		for {
			select {
			case m, ok := <-matchChan:
				if !ok {
					matchChan = nil
				} else {
					results = append(results, m)
				}
			case err, ok := <-errChan:
				if !ok {
					errChan = nil
				} else if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
			if matchChan == nil && errChan == nil {
				break
			}
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 matches for similarity search, got %d", len(results))
		}
		
		// test1.java should have a lower score (closer) than test2.java
		var score1, score2 float64
		for _, r := range results {
			if filepath.Base(r.Path) == "test1.java" {
				score1 = r.Score
			} else if filepath.Base(r.Path) == "test2.java" {
				score2 = r.Score
			}
		}

		if score1 >= score2 {
			t.Errorf("Expected test1.java (score %v) to be more similar than test2.java (score %v)", score1, score2)
		}
	})

	t.Run("Ignore Pattern", func(t *testing.T) {
		s.Ignore = []string{"test2"}
		ctx := context.Background()
		matchChan, errChan := s.Search(ctx, tmpDir, ".*")

		count := 0
		for {
			select {
			case m, ok := <-matchChan:
				if !ok {
					matchChan = nil
				} else {
					if filepath.Base(m.Path) == "test2.java" {
						t.Errorf("test2.java should have been ignored")
					}
					count++
				}
			case _, ok := <-errChan:
				if !ok {
					errChan = nil
				}
			}
			if matchChan == nil && errChan == nil {
				break
			}
		}
	})
}
