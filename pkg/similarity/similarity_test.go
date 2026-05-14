package similarity

import (
	"testing"
)

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		want     float64
	}{
		{"Empty strings", "", "", 0},
		{"Identical strings", "hello", "hello", 0},
		{"Single character difference", "hello", "hallo", 1},
		{"Completely different", "abc", "def", 3},
		{"One empty string", "abc", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateDistance(tt.s1, tt.s2); got != tt.want {
				t.Errorf("CalculateDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
