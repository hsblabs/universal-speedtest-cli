package stats

import (
	"math"
	"sort"
)

// Average returns the arithmetic mean of the given values.
func Average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Median returns the middle value of the sorted values.
func Median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	half := len(sorted) / 2
	if len(sorted)%2 != 0 {
		return sorted[half]
	}
	return (sorted[half-1] + sorted[half]) / 2.0
}

// Quartile returns the value at the given percentile (0.0–1.0) using linear interpolation.
func Quartile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	pos := float64(len(sorted)-1) * percentile
	base := int(math.Floor(pos))
	rest := pos - float64(base)
	if base+1 < len(sorted) {
		return sorted[base] + rest*(sorted[base+1]-sorted[base])
	}
	return sorted[base]
}

// Jitter returns the average absolute difference between consecutive values.
func Jitter(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	diffs := make([]float64, 0, len(values)-1)
	for i := 0; i < len(values)-1; i++ {
		diffs = append(diffs, math.Abs(values[i]-values[i+1]))
	}
	return Average(diffs)
}
