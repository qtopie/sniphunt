package search

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/qtopie/sniphunt/pkg/similarity"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024*1024) // 1MB buffer
	},
}

func (s *Searcher) regexWorker(ctx context.Context, pattern *regexp.Regexp, literal string, tasks <-chan task, results chan<- Match) error {
	literalBytes := []byte(literal)
	for t := range tasks {
		if s.SearchZips && (strings.HasSuffix(t.path, ".zip") || strings.HasSuffix(t.path, ".jar")) {
			s.searchZip(ctx, t.path, pattern, literalBytes, results)
			continue
		}

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
			if len(literalBytes) > 0 && !bytes.Contains(line, literalBytes) {
				continue
			}

			if pattern.Match(line) {
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

func (s *Searcher) searchZip(ctx context.Context, path string, pattern *regexp.Regexp, literal []byte, results chan<- Match) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(rc)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Bytes()
			if len(literal) > 0 && !bytes.Contains(line, literal) {
				continue
			}
			if pattern.Match(line) {
				select {
				case results <- Match{
					Path:    path + "::" + f.Name,
					LineNum: lineNum,
					Text:    append([]byte{}, line...),
				}:
				case <-ctx.Done():
					rc.Close()
					return
				}
			}
		}
		rc.Close()
	}
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
