package search

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
	"github.com/masatana/sniphunt/pkg/similarity"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 64*1024) // 64KB buffer
	},
}

func (s *Searcher) regexWorker(ctx context.Context, pattern *regexp2.Regexp, literal string, tasks <-chan task, results chan<- Match) error {
	for t := range tasks {
		file, err := os.Open(t.path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		buf := bufferPool.Get().([]byte)
		scanner.Buffer(buf, len(buf))

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Bytes()

			// Pre-filter with literal string if available
			if literal != "" && !strings.Contains(string(line), literal) {
				continue
			}

			match, err := pattern.MatchString(string(line))
			if err == nil && match {
				select {
				case results <- Match{
					Path:    t.path,
					LineNum: lineNum,
					Text:    append([]byte{}, line...), // Copy line
				}:
				case <-ctx.Done():
					file.Close()
					bufferPool.Put(buf)
					return ctx.Err()
				}
			}
		}
		file.Close()
		bufferPool.Put(buf)
	}
	return nil
}

func (s *Searcher) similarityWorker(ctx context.Context, targetSnippet string, tasks <-chan task, results chan<- Match) error {
	for t := range tasks {
		content, err := os.ReadFile(t.path)
		if err != nil {
			continue
		}

		score := similarity.CalculateDistance(targetSnippet, string(content))
		select {
		case results <- Match{
			Path:  t.path,
			Score: score,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
