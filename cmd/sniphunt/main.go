package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/masatana/sniphunt/pkg/search"
)

var (
	inputFile   string
	pattern     string
	searchDir   string
	workers     int
	ignorePaths string
	extensions  string
)

func main() {
	flag.StringVar(&inputFile, "input", "", "input code snippet file (for similarity search)")
	flag.StringVar(&pattern, "pattern", "", "regex pattern (for regex search)")
	flag.StringVar(&searchDir, "dir", ".", "search directory")
	flag.IntVar(&workers, "workers", 0, "number of workers (default: NumCPU)")
	flag.StringVar(&ignorePaths, "ignore", "", "comma-separated list of paths to ignore")
	flag.StringVar(&extensions, "ext", ".java", "comma-separated list of extensions to include")
	flag.Parse()

	if inputFile == "" && pattern == "" {
		fmt.Println("Error: either -input or -pattern must be specified")
		flag.Usage()
		os.Exit(1)
	}

	s := search.NewSearcher()
	if workers > 0 {
		s.Workers = workers
	}
	if ignorePaths != "" {
		s.Ignore = strings.Split(ignorePaths, ",")
	}
	if extensions != "" {
		s.Extensions = strings.Split(extensions, ",")
	}

	ctx := context.Background()

	if pattern != "" {
		runRegexSearch(ctx, s, searchDir, pattern)
	} else if inputFile != "" {
		runSimilaritySearch(ctx, s, searchDir, inputFile)
	}
}

func runRegexSearch(ctx context.Context, s *search.Searcher, root, pattern string) {
	matchChan, errChan := s.Search(ctx, root, pattern)

	for {
		select {
		case match, ok := <-matchChan:
			if !ok {
				matchChan = nil
			} else {
				fmt.Printf("%s:%d: %s\n", match.Path, match.LineNum, strings.TrimSpace(string(match.Text)))
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else if err != nil {
				log.Printf("Error during search: %v", err)
			}
		}

		if matchChan == nil && errChan == nil {
			break
		}
	}
}

func runSimilaritySearch(ctx context.Context, s *search.Searcher, root, inputFile string) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	snippet := string(data)
	matchChan, errChan := s.MatchSnippet(ctx, root, snippet)

	var results []search.Match
	for {
		select {
		case match, ok := <-matchChan:
			if !ok {
				matchChan = nil
			} else {
				results = append(results, match)
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else if err != nil {
				log.Printf("Error during search: %v", err)
			}
		}

		if matchChan == nil && errChan == nil {
			break
		}
	}

	if len(results) == 0 {
		fmt.Println("No matches found.")
		return
	}

	// Find the most similar one (lowest distance)
	var closest search.Match
	first := true
	for _, r := range results {
		if first || r.Score < closest.Score {
			closest = r
			first = false
		}
	}

	fmt.Printf("Closest match: %s (Distance: %.2f)\n", closest.Path, closest.Score)
}
