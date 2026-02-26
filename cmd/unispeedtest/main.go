package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/hsblabs/universal-speedtest-cli/internal/cloudflare"
	"github.com/hsblabs/universal-speedtest-cli/internal/color"
	"github.com/hsblabs/universal-speedtest-cli/internal/reporter"
	"github.com/hsblabs/universal-speedtest-cli/internal/stats"
)

func main() {
	jsonOut := flag.Bool("json", false, "Output results in JSON format")
	prettyOut := flag.Bool("pretty", false, "Output pretty-printed JSON (implies -json)")
	flag.Parse()

	if *prettyOut {
		*jsonOut = true
	}

	verbose := !*jsonOut

	var progress func(string, ...interface{})
	if verbose {
		progress = func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
		}
	} else {
		progress = func(string, ...interface{}) {}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nInterrupted.")
		os.Exit(130)
	}()

	progress("%sInitializing Cloudflare Speed Test...%s\n\n", color.Bold, color.Reset)

	meta, err := cloudflare.FetchMeta()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not fetch network metadata: %v\n", err)
	}

	unloadedLatencies := cloudflare.MeasureLatency()
	unloadedMed := stats.Median(unloadedLatencies)
	unloadedJitter := stats.Jitter(unloadedLatencies)

	progress("%s[ Download Measurements ]%s\n", color.Bold, color.Reset)
	var progressWriter *os.File
	if verbose {
		progressWriter = os.Stdout
	}
	testSizes := []int{101000, 1001000, 10001000, 25001000}
	downCounts := []int{10, 8, 6, 4}
	downSpeeds, downLatencies := cloudflare.MeasurePhase("download", testSizes, downCounts, progressWriter)
	downOverall := stats.Quartile(downSpeeds, 0.90)

	progress("\n%s[ Upload Measurements ]%s\n", color.Bold, color.Reset)
	upCounts := []int{8, 6, 4, 4}
	upSpeeds, upLatencies := cloudflare.MeasurePhase("upload", testSizes, upCounts, progressWriter)
	upOverall := stats.Quartile(upSpeeds, 0.90)

	progress("\n%s[ Packet Loss Test ]%s Running 1000 requests...\n", color.Bold, color.Reset)
	lossPercent, received, total := cloudflare.MeasurePacketLoss()

	result := reporter.Result{
		DownloadMbps:      downOverall,
		UploadMbps:        upOverall,
		UnloadedLatency:   unloadedMed,
		LoadedDownLatency: stats.Median(downLatencies),
		LoadedUpLatency:   stats.Median(upLatencies),
		Jitter:            unloadedJitter,
		PacketLoss:        lossPercent,
		Received:          received,
		Total:             total,
		ServerColo:        meta.Colo.City,
		NetworkASN:        fmt.Sprintf("AS%d", meta.ASN),
		NetworkASOrg:      meta.ASOrganization,
		IP:                meta.ClientIP,
	}

	if verbose {
		reporter.PrintHuman(os.Stdout, result)
	} else {
		if err := reporter.PrintJSON(os.Stdout, result, *prettyOut); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	}
}
