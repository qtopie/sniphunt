package similarity

import textdistance "github.com/masatana/go-textdistance"

// CalculateDistance returns the Levenshtein distance between two strings.
func CalculateDistance(s1, s2 string) float64 {
	return float64(textdistance.LevenshteinDistance(s1, s2))
}
