package cloudflare_test

import (
	"math"
	"testing"

	"github.com/hsblabs/universal-speedtest-cli/internal/cloudflare"
)

func TestMeasureSpeed(t *testing.T) {
	tests := []struct {
		name       string
		byteCount  int
		durationMs float64
		expectMbps float64
	}{
		{
			name:       "1MB in 1000ms = 8 Mbps",
			byteCount:  1_000_000,
			durationMs: 1000,
			expectMbps: 8.0,
		},
		{
			name:       "10MB in 1000ms = 80 Mbps",
			byteCount:  10_000_000,
			durationMs: 1000,
			expectMbps: 80.0,
		},
		{
			name:       "100MB in 1000ms = 800 Mbps",
			byteCount:  100_000_000,
			durationMs: 1000,
			expectMbps: 800.0,
		},
		{
			name:       "1MB in 500ms = 16 Mbps",
			byteCount:  1_000_000,
			durationMs: 500,
			expectMbps: 16.0,
		},
		{
			name:       "100kB in 1000ms = 0.8 Mbps",
			byteCount:  100_000,
			durationMs: 1000,
			expectMbps: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cloudflare.MeasureSpeed(tt.byteCount, tt.durationMs)
			if math.Abs(got-tt.expectMbps) > 1e-6 {
				t.Errorf("MeasureSpeed(%d, %v) = %v, want %v", tt.byteCount, tt.durationMs, got, tt.expectMbps)
			}
		})
	}
}
