package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hsblabs/universal-speedtest-cli/internal/color"
)

// Result holds all measurements from a speed test run.
type Result struct {
	DownloadMbps    float64
	UploadMbps      float64
	UnloadedLatency float64
	LoadedDownLatency float64
	LoadedUpLatency float64
	Jitter          float64
	PacketLoss      float64
	Received        int
	Total           int
	ServerColo      string
	NetworkASN      string
	NetworkASOrg    string
	IP              string
}

// PrintHuman writes a human-readable speed test report to w.
func PrintHuman(w io.Writer, r Result) {
	score := EvaluateQuality(r.DownloadMbps, r.UploadMbps, r.UnloadedLatency, r.Jitter, r.PacketLoss)

	fmt.Fprintf(w, "\n======================================================\n")
	fmt.Fprintf(w, "%s         YOUR INTERNET SPEED REPORT%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "======================================================\n\n")

	fmt.Fprintf(w, "%sOverall Download:%s %s%.2f Mbps%s\n", color.Bold, color.Reset, color.Green, r.DownloadMbps, color.Reset)
	fmt.Fprintf(w, "%sOverall Upload:%s   %s%.2f Mbps%s\n\n", color.Bold, color.Reset, color.Green, r.UploadMbps, color.Reset)

	fmt.Fprintf(w, "%sLatency & Jitter:%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "  Unloaded Latency: %s%.2f ms%s (Jitter: %s%.2f ms%s)\n",
		color.Magenta, r.UnloadedLatency, color.Reset,
		color.Magenta, r.Jitter, color.Reset)
	fmt.Fprintf(w, "  Down-Loaded Latency: %.2f ms\n", r.LoadedDownLatency)
	fmt.Fprintf(w, "  Up-Loaded Latency: %.2f ms\n\n", r.LoadedUpLatency)

	fmt.Fprintf(w, "%sPacket Loss:%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "  %.1f%% (Received %d / %d)\n\n", r.PacketLoss, r.Received, r.Total)

	fmt.Fprintf(w, "%sNetwork Quality Score:%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "  Video Streaming: %s  |  Online Gaming: %s  |  Video Chatting: %s\n\n",
		qualityLabel(score.Streaming),
		qualityLabel(score.Gaming),
		qualityLabel(score.Chatting),
	)

	fmt.Fprintf(w, "%sServer & Connection:%s\n", color.Bold, color.Reset)
	fmt.Fprintf(w, "  Server Location: %s\n", r.ServerColo)
	fmt.Fprintf(w, "  Your Network:    %s (%s)\n", r.NetworkASOrg, r.NetworkASN)
	fmt.Fprintf(w, "  Your IP address: %s\n", r.IP)
	fmt.Fprintf(w, "======================================================\n")
}

type jsonLatency struct {
	Unloaded   float64 `json:"unloaded"`
	LoadedDown float64 `json:"loaded_down"`
	LoadedUp   float64 `json:"loaded_up"`
	Jitter     float64 `json:"jitter"`
}

type jsonResult struct {
	DownloadMbps      float64     `json:"download_mbps"`
	UploadMbps        float64     `json:"upload_mbps"`
	LatencyMs         jsonLatency `json:"latency_ms"`
	PacketLossPercent float64     `json:"packet_loss_percent"`
	ServerColo        string      `json:"server_colo"`
	NetworkASN        string      `json:"network_asn"`
	NetworkASOrg      string      `json:"network_as_org"`
	IP                string      `json:"ip"`
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
