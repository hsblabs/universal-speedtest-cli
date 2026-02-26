package stats_test

import (
	"math"
	"testing"

	"github.com/hsblabs/universal-speedtest-cli/internal/stats"
)

func TestAverage(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		expect float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"multiple", []float64{1, 2, 3, 4, 5}, 3},
		{"with decimals", []float64{1.5, 2.5, 3.0}, 7.0 / 3.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stats.Average(tt.input)
			if math.Abs(got-tt.expect) > 1e-9 {
				t.Errorf("Average(%v) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		expect float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{7}, 7},
		{"odd count", []float64{3, 1, 2}, 2},
		{"even count", []float64{4, 1, 3, 2}, 2.5},
		{"already sorted", []float64{10, 20, 30}, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stats.Median(tt.input)
			if math.Abs(got-tt.expect) > 1e-9 {
				t.Errorf("Median(%v) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestQuartile(t *testing.T) {
	tests := []struct {
		name       string
		input      []float64
		percentile float64
		expect     float64
	}{
		{"empty", []float64{}, 0.5, 0},
		{"single", []float64{10}, 0.5, 10},
		{"p50 of 5 values", []float64{1, 2, 3, 4, 5}, 0.5, 3},
		{"p90 of 10 values", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 0.9, 9.1},
		{"p0 returns min", []float64{5, 3, 8, 1}, 0.0, 1},
		{"p1 returns max", []float64{5, 3, 8, 1}, 1.0, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stats.Quartile(tt.input, tt.percentile)
			if math.Abs(got-tt.expect) > 1e-9 {
				t.Errorf("Quartile(%v, %v) = %v, want %v", tt.input, tt.percentile, got, tt.expect)
			}
		})
	}
}

func TestJitter(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		expect float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 0},
		{"two equal values", []float64{5, 5}, 0},
		{"two values", []float64{10, 20}, 10},
		{"three values", []float64{10, 20, 10}, 10},
		{"stable signal", []float64{5, 5, 5, 5}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stats.Jitter(tt.input)
			if math.Abs(got-tt.expect) > 1e-9 {
				t.Errorf("Jitter(%v) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestMedianDoesNotMutateInput(t *testing.T) {
	input := []float64{5, 3, 1, 4, 2}
	original := make([]float64, len(input))
	copy(original, input)
	stats.Median(input)
	for i, v := range input {
		if v != original[i] {
			t.Errorf("Median mutated input at index %d: got %v, want %v", i, v, original[i])
		}
	}
}
