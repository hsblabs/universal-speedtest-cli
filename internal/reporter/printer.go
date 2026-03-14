package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hsblabs/universal-speedtest-cli/internal/color"
)

// Result holds all measurements from a speed test run.
type Result struct {
	DownloadMbps      *float64
	UploadMbps        *float64
	UnloadedLatency   *float64
	LoadedDownLatency *float64
	LoadedUpLatency   *float64
	Jitter            *float64
	PacketLoss        *float64
	Received          int
	Total             int
	ServerColo        *string
	NetworkASN        *string
	NetworkASOrg      *string
	IP                *string
	Warnings          []string
}

// PrintHuman writes a human-readable speed test report to w.
func PrintHuman(w io.Writer, r Result) {
	score, ok := evaluateQualityScore(r)

	fmt.Fprintf(w, "\n======================================================\n")
	fmt.Fprintf(w, "%s         YOUR INTERNET SPEED REPORT%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "======================================================\n\n")

	fmt.Fprintf(w, "%sOverall Download:%s %s\n", color.Bold, color.Reset, formatMetric(r.DownloadMbps, "Mbps", color.Green))
	fmt.Fprintf(w, "%sOverall Upload:%s   %s\n\n", color.Bold, color.Reset, formatMetric(r.UploadMbps, "Mbps", color.Green))

	fmt.Fprintf(w, "%sLatency & Jitter:%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "  Unloaded Latency: %s (Jitter: %s)\n",
		formatMetric(r.UnloadedLatency, "ms", color.Magenta),
		formatMetric(r.Jitter, "ms", color.Magenta),
	)
	fmt.Fprintf(w, "  Down-Loaded Latency: %s\n", formatMetric(r.LoadedDownLatency, "ms", ""))
	fmt.Fprintf(w, "  Up-Loaded Latency: %s\n\n", formatMetric(r.LoadedUpLatency, "ms", ""))

	fmt.Fprintf(w, "%sPacket Loss:%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "  %s (Received %d / %d)\n\n", formatPercent(r.PacketLoss), r.Received, r.Total)

	fmt.Fprintf(w, "%sNetwork Quality Score:%s\n", color.Bold, color.Reset)
	if ok {
		fmt.Fprintf(w, "  Video Streaming: %s  |  Online Gaming: %s  |  Video Chatting: %s\n\n",
			qualityLabel(score.Streaming),
			qualityLabel(score.Gaming),
			qualityLabel(score.Chatting),
		)
	} else {
		fmt.Fprintf(w, "  N/A (insufficient data)\n\n")
	}

	fmt.Fprintf(w, "%sServer & Connection:%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "  Server Location: %s\n", displayString(r.ServerColo))
	fmt.Fprintf(w, "  Your Network:    %s (%s)\n", displayString(r.NetworkASOrg), displayString(r.NetworkASN))
	fmt.Fprintf(w, "  Your IP address: %s\n", displayString(r.IP))
	if len(r.Warnings) > 0 {
		fmt.Fprintf(w, "\n%sWarnings:%s\n", color.Bold, color.Reset)
		for _, warning := range r.Warnings {
			fmt.Fprintf(w, "  - %s%s%s\n", color.Yellow, warning, color.Reset)
		}
	}
	fmt.Fprintf(w, "======================================================\n")
}

type jsonLatency struct {
	Unloaded   *float64 `json:"unloaded"`
	LoadedDown *float64 `json:"loaded_down"`
	LoadedUp   *float64 `json:"loaded_up"`
	Jitter     *float64 `json:"jitter"`
}

type jsonResult struct {
	DownloadMbps      *float64    `json:"download_mbps"`
	UploadMbps        *float64    `json:"upload_mbps"`
	LatencyMs         jsonLatency `json:"latency_ms"`
	PacketLossPercent *float64    `json:"packet_loss_percent"`
	ServerColo        *string     `json:"server_colo"`
	NetworkASN        *string     `json:"network_asn"`
	NetworkASOrg      *string     `json:"network_as_org"`
	IP                *string     `json:"ip"`
	Warnings          []string    `json:"warnings,omitempty"`
}

// PrintJSON writes the result as a JSON object to w. Set pretty=true for indented output.
func PrintJSON(w io.Writer, r Result, pretty bool) error {
	data := jsonResult{
		DownloadMbps: r.DownloadMbps,
		UploadMbps:   r.UploadMbps,
		LatencyMs: jsonLatency{
			Unloaded:   r.UnloadedLatency,
			LoadedDown: r.LoadedDownLatency,
			LoadedUp:   r.LoadedUpLatency,
			Jitter:     r.Jitter,
		},
		PacketLossPercent: r.PacketLoss,
		ServerColo:        r.ServerColo,
		NetworkASN:        r.NetworkASN,
		NetworkASOrg:      r.NetworkASOrg,
		IP:                r.IP,
		Warnings:          append([]string(nil), r.Warnings...),
	}

	var out []byte
	var err error
	if pretty {
		out, err = json.MarshalIndent(data, "", "  ")
	} else {
		out, err = json.Marshal(data)
	}
	if err != nil {
		return err
	}

	fmt.Fprintln(w, string(out))
	return nil
}

func qualityLabel(good bool) string {
	if good {
		return color.Green + "Good" + color.Reset
	}
	return color.Red + "Poor" + color.Reset
}

func evaluateQualityScore(r Result) (QualityScore, bool) {
	if r.DownloadMbps == nil || r.UploadMbps == nil || r.UnloadedLatency == nil || r.Jitter == nil || r.PacketLoss == nil {
		return QualityScore{}, false
	}

	return EvaluateQuality(*r.DownloadMbps, *r.UploadMbps, *r.UnloadedLatency, *r.Jitter, *r.PacketLoss), true
}

func formatMetric(value *float64, unit string, colorCode string) string {
	if value == nil {
		return "N/A"
	}

	rendered := fmt.Sprintf("%.2f %s", *value, unit)
	if colorCode == "" {
		return rendered
	}

	return colorCode + rendered + color.Reset
}

func formatPercent(value *float64) string {
	if value == nil {
		return "N/A"
	}

	return fmt.Sprintf("%.1f%%", *value)
}

func displayString(value *string) string {
	if value == nil || *value == "" {
		return "N/A"
	}

	return *value
}
