package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/hsblabs/universal-speedtest-cli/internal/cloudflare"
	"github.com/hsblabs/universal-speedtest-cli/internal/color"
	"github.com/hsblabs/universal-speedtest-cli/internal/reporter"
	"github.com/hsblabs/universal-speedtest-cli/internal/stats"
)

var (
	version         = ""
	buildInfoReader = debug.ReadBuildInfo
	benchmarkMain   = func(jsonOut, prettyOut bool, stdout, stderr io.Writer) int {
		if prettyOut {
			jsonOut = true
		}

		verbose := !jsonOut

		var progress func(string, ...interface{})
		if verbose {
			progress = func(format string, args ...interface{}) {
				fmt.Fprintf(stdout, format, args...)
			}
		} else {
			progress = func(string, ...interface{}) {}
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer signal.Stop(sigCh)

		go func() {
			<-sigCh
			fmt.Fprintln(stderr, "\nInterrupted.")
			// Intentionally exit the whole process on SIGINT (130) instead of
			// returning through run(). Signal handling is process-wide and an
			// immediate exit avoids leaving background goroutines or partial state.
			os.Exit(130)
		}()

		progress("%sInitializing Cloudflare Speed Test...%s\n\n", color.Bold, color.Reset)

		meta, err := cloudflare.FetchMeta()
		if err != nil {
			fmt.Fprintf(stderr, "Warning: could not fetch network metadata: %v\n", err)
		}

		unloadedLatencies := cloudflare.MeasureLatency()
		unloadedMed := stats.Median(unloadedLatencies)
		unloadedJitter := stats.Jitter(unloadedLatencies)

		progress("%s[ Download Measurements ]%s\n", color.Bold, color.Reset)
		var progressWriter io.Writer
		if verbose {
			progressWriter = stdout
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
			reporter.PrintHuman(stdout, result)
			return 0
		}

		if err := reporter.PrintJSON(stdout, result, prettyOut); err != nil {
			fmt.Fprintf(stderr, "error encoding JSON: %v\n", err)
			return 1
		}

		return 0
	}
)

func hasVersionFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			return true
		}
	}

	return false
}

func resolvedVersion() string {
	if version != "" {
		return version
	}

	info, ok := buildInfoReader()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "dev"
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if hasVersionFlag(args) {
		fmt.Fprintln(stdout, resolvedVersion())
		return 0
	}

	fs := flag.NewFlagSet("unispeedtest", flag.ContinueOnError)
	fs.SetOutput(stderr)

	jsonOut := fs.Bool("json", false, "Output results in JSON format")
	prettyOut := fs.Bool("pretty", false, "Output pretty-printed JSON (implies -json)")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	return benchmarkMain(*jsonOut, *prettyOut, stdout, stderr)
}
