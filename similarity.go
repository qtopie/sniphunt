package main

import textdistance "github.com/masatana/go-textdistance"

var (
	snippet         string
	similarSnippets map[string]float64 = make(map[string]float64)
)

type Snippet struct {
	Path       string
	Similarity float64
}

type Snippets []Snippet

func (s Snippets) Len() int { return len(s) }

func (s Snippets) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s Snippets) Less(i, j int) bool {
	return s[i].Similarity < s[j].Similarity
}

func caclDistance(s1, s2 string) float64 {
	return float64(textdistance.LevenshteinDistance(s1, s2))
}
