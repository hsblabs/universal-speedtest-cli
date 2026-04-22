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

		var result reporter.Result
		var warnings []string

		meta, err := cloudflare.FetchMeta()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("network metadata unavailable: %v", err))
		} else {
			result.ServerColo = stringPtr(meta.Colo.City)
			result.NetworkASN = stringPtr(fmt.Sprintf("AS%d", meta.ASN))
			result.NetworkASOrg = stringPtr(meta.ASOrganization)
			result.IP = stringPtr(meta.ClientIP)
		}

		latencyMeasurement := cloudflare.MeasureLatency()
		warnings = append(warnings, latencyMeasurement.Warnings...)
		if len(latencyMeasurement.Samples) == 0 {
			fmt.Fprintln(stderr, "error: unloaded latency measurement produced no successful samples")
			return 1
		}
		result.UnloadedLatency = float64Ptr(stats.Median(latencyMeasurement.Samples))
		if len(latencyMeasurement.Samples) > 1 {
			result.Jitter = float64Ptr(stats.Jitter(latencyMeasurement.Samples))
		} else {
			warnings = append(warnings, "jitter unavailable: need at least 2 unloaded latency samples")
		}

		progress("%s[ Download Measurements ]%s\n", color.Bold, color.Reset)
		var progressWriter io.Writer
		if verbose {
			progressWriter = stdout
		}
		downloadSpecs := []cloudflare.PhaseSpec{
			{SizeBytes: 101000, Count: 10},
			{SizeBytes: 1001000, Count: 8},
			{SizeBytes: 10001000, Count: 6},
			{SizeBytes: 25001000, Count: 4},
		}
		downloadMeasurement, err := cloudflare.MeasurePhase("download", downloadSpecs, progressWriter)
		if err != nil {
			fmt.Fprintf(stderr, "error: download measurement failed: %v\n", err)
			return 1
		}
		warnings = append(warnings, downloadMeasurement.Warnings...)
		if len(downloadMeasurement.Speeds) == 0 {
			fmt.Fprintln(stderr, "error: download measurement produced no successful samples")
			return 1
		}
		result.DownloadMbps = float64Ptr(stats.Quartile(downloadMeasurement.Speeds, 0.90))
		if len(downloadMeasurement.LoadedLatencies) > 0 {
			result.LoadedDownLatency = float64Ptr(stats.Median(downloadMeasurement.LoadedLatencies))
		} else {
			warnings = append(warnings, "download loaded latency unavailable: no samples collected")
		}

		progress("\n%s[ Upload Measurements ]%s\n", color.Bold, color.Reset)
		uploadSpecs := []cloudflare.PhaseSpec{
			{SizeBytes: 101000, Count: 8},
			{SizeBytes: 1001000, Count: 6},
			{SizeBytes: 10001000, Count: 4},
			{SizeBytes: 25001000, Count: 4},
		}
		uploadMeasurement, err := cloudflare.MeasurePhase("upload", uploadSpecs, progressWriter)
		if err != nil {
			fmt.Fprintf(stderr, "error: upload measurement failed: %v\n", err)
			return 1
		}
		warnings = append(warnings, uploadMeasurement.Warnings...)
		if len(uploadMeasurement.Speeds) == 0 {
			fmt.Fprintln(stderr, "error: upload measurement produced no successful samples")
			return 1
		}
		result.UploadMbps = float64Ptr(stats.Quartile(uploadMeasurement.Speeds, 0.90))
		if len(uploadMeasurement.LoadedLatencies) > 0 {
			result.LoadedUpLatency = float64Ptr(stats.Median(uploadMeasurement.LoadedLatencies))
		} else {
			warnings = append(warnings, "upload loaded latency unavailable: no samples collected")
		}

		progress("\n%s[ Packet Loss Test ]%s Running 1000 requests...\n", color.Bold, color.Reset)
		lossPercent, received, total := cloudflare.MeasurePacketLoss()

		result.PacketLoss = float64Ptr(lossPercent)
		result.Received = received
		result.Total = total
		result.Warnings = warnings

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
	if ok && info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
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

func float64Ptr(value float64) *float64 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
