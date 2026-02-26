package reporter_test

import (
	"testing"

	"github.com/hsblabs/universal-speedtest-cli/internal/reporter"
)

func TestEvaluateQuality(t *testing.T) {
	tests := []struct {
		name       string
		download   float64
		upload     float64
		latency    float64
		jitter     float64
		packetLoss float64
		want       reporter.QualityScore
	}{
		{
			name:       "excellent connection",
			download:   100,
			upload:     50,
			latency:    10,
			jitter:     2,
			packetLoss: 0,
			want:       reporter.QualityScore{Streaming: true, Gaming: true, Chatting: true},
		},
		{
			name:       "high latency fails gaming",
			download:   100,
			upload:     50,
			latency:    60,
			jitter:     5,
			packetLoss: 0,
			want:       reporter.QualityScore{Streaming: true, Gaming: false, Chatting: true},
		},
		{
			name:       "high jitter fails gaming but not chatting",
			download:   100,
			upload:     50,
			latency:    20,
			jitter:     25,
			packetLoss: 0,
			// jitter=25 fails gaming (<20) but passes chatting (<30)
			want: reporter.QualityScore{Streaming: true, Gaming: false, Chatting: true},
		},
		{
			name:       "slow download fails streaming",
			download:   3,
			upload:     10,
			latency:    20,
			jitter:     5,
			packetLoss: 0,
			want:       reporter.QualityScore{Streaming: false, Gaming: true, Chatting: true},
		},
		{
			name:       "slow download fails chatting",
			download:   1,
			upload:     10,
			latency:    20,
			jitter:     5,
			packetLoss: 0,
			want:       reporter.QualityScore{Streaming: false, Gaming: true, Chatting: false},
		},
		{
			name:       "slow upload fails chatting",
			download:   100,
			upload:     1,
			latency:    20,
			jitter:     5,
			packetLoss: 0,
			want:       reporter.QualityScore{Streaming: true, Gaming: true, Chatting: false},
		},
		{
			name:       "high packet loss fails all",
			download:   100,
			upload:     50,
			latency:    10,
			jitter:     2,
			packetLoss: 5,
			want:       reporter.QualityScore{Streaming: false, Gaming: false, Chatting: false},
		},
		{
			name:       "packet loss below streaming but above chatting threshold",
			download:   10,
			upload:     10,
			latency:    50,
			jitter:     5,
			packetLoss: 1.9,
			// packetLoss=1.9 passes streaming (<2.0) but fails chatting (<1.0)
			want: reporter.QualityScore{Streaming: true, Gaming: false, Chatting: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reporter.EvaluateQuality(tt.download, tt.upload, tt.latency, tt.jitter, tt.packetLoss)
			if got != tt.want {
				t.Errorf("EvaluateQuality() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
