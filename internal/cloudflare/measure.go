package cloudflare

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/hsblabs/universal-speedtest-cli/internal/stats"
)

const unloadedLatencySampleCount = 20

type sampleFailureTracker struct {
	count int
	first error
}

func (t *sampleFailureTracker) Record(err error) {
	if err == nil {
		return
	}

	t.count++
	if t.first == nil {
		t.first = err
	}
}

func (t sampleFailureTracker) Message(format string) []string {
	if t.count == 0 {
		return nil
	}

	return []string{
		fmt.Sprintf(format, t.count, t.first),
	}
}

// MeasureSpeed converts a transfer of byteCount bytes completed in durationMs milliseconds
// into megabits per second.
func MeasureSpeed(byteCount int, durationMs float64) float64 {
	return (float64(byteCount*8) / (durationMs / 1000)) / 1e6
}

func latencySample() (float64, error) {
	pd, err := MakeRequest("GET", "/__down?bytes=0", nil)
	if err != nil {
		return 0, err
	}
	if pd.TTFB.IsZero() {
		return 0, errors.New("missing first response byte timing")
	}

	serverTiming, err := ParseServerTiming(pd.ServerTimingHeader)
	if err != nil {
		return 0, err
	}

	dur := pd.TTFB.Sub(pd.Started).Seconds()*1000 - serverTiming
	if dur <= 0 {
		return 0, fmt.Errorf("computed non-positive latency: %.2f ms", dur)
	}

	return dur, nil
}

// MeasureLatency sends lightweight requests and returns the unloaded latency samples in ms.
func MeasureLatency() LatencyMeasurement {
	var measurements []float64

	var failures sampleFailureTracker
	for i := 0; i < unloadedLatencySampleCount; i++ {
		sample, err := latencySample()
		if err != nil {
			failures.Record(err)
			continue
		}

		measurements = append(measurements, sample)
	}

	return LatencyMeasurement{
		Samples:       measurements,
		FailedSamples: failures.count,
		Warnings:      failures.Message("unloaded latency measurement failed %d time(s); first error: %v"),
	}
}

func downloadSample(size int) (float64, error) {
	pd, err := MakeRequest("GET", fmt.Sprintf("/__down?bytes=%d", size), nil)
	if err != nil {
		return 0, err
	}
	if pd.TTFB.IsZero() {
		return 0, errors.New("missing first response byte timing")
	}

	transferTime := pd.Ended.Sub(pd.TTFB).Seconds() * 1000
	if transferTime <= 0 {
		return 0, fmt.Errorf("computed non-positive transfer time: %.2f ms", transferTime)
	}

	return MeasureSpeed(size, transferTime), nil
}

func uploadSample(size int, payload []byte) (float64, error, error) {
	pd, err := MakeRequest("POST", "/__up", payload)
	if err != nil {
		return 0, nil, err
	}

	serverTiming, err := ParseServerTiming(pd.ServerTimingHeader)
	if err == nil && serverTiming > 0 {
		return MeasureSpeed(size, serverTiming), nil, nil
	}

	fallbackDuration := pd.Ended.Sub(pd.Started).Seconds() * 1000
	if fallbackDuration <= 0 {
		if err != nil {
			return 0, nil, err
		}
		return 0, nil, fmt.Errorf("computed non-positive upload duration: %.2f ms", fallbackDuration)
	}

	if err != nil {
		return MeasureSpeed(size, fallbackDuration), fmt.Errorf("upload server timing unavailable, fell back to end-to-end duration: %w", err), nil
	}

	return MeasureSpeed(size, fallbackDuration), fmt.Errorf("upload server timing was non-positive (%.2f ms), fell back to end-to-end duration", serverTiming), nil
}

func validatePhaseSpecs(phaseType string, specs []PhaseSpec) error {
	switch phaseType {
	case "download", "upload":
	default:
		return fmt.Errorf("unsupported phase type %q", phaseType)
	}

	if len(specs) == 0 {
		return errors.New("at least one phase spec is required")
	}

	for _, spec := range specs {
		if spec.SizeBytes <= 0 {
			return fmt.Errorf("phase spec size must be positive: %d", spec.SizeBytes)
		}
		if spec.Count <= 0 {
			return fmt.Errorf("phase spec count must be positive for size %d: %d", spec.SizeBytes, spec.Count)
		}
	}

	return nil
}

// backgroundLatencyMonitor periodically samples latency while other measurements run.
func backgroundLatencyMonitor(ctx context.Context, results *[]float64, failures *sampleFailureTracker) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sample, err := latencySample()
			if err != nil {
				failures.Record(err)
				continue
			}

			*results = append(*results, sample)
		}
	}
}

// MeasurePhase runs either a download or upload measurement phase across the given sizes and
// sample specs. Progress is written to w if non-nil.
func MeasurePhase(phaseType string, specs []PhaseSpec, w io.Writer) (PhaseMeasurement, error) {
	if err := validatePhaseSpecs(phaseType, specs); err != nil {
		return PhaseMeasurement{}, err
	}

	var allSpeeds []float64
	var loadedLatencies []float64
	var throughputFailures sampleFailureTracker
	var throughputWarnings sampleFailureTracker
	var loadedLatencyFailures sampleFailureTracker

	ctx, cancel := context.WithCancel(context.Background())
	var monitorWG sync.WaitGroup
	monitorWG.Add(1)
	go func() {
		defer monitorWG.Done()
		backgroundLatencyMonitor(ctx, &loadedLatencies, &loadedLatencyFailures)
	}()

	for _, spec := range specs {
		var speeds []float64

		var payload []byte
		if phaseType != "download" {
			payload = bytes.Repeat([]byte("0"), spec.SizeBytes)
		}

		for j := 0; j < spec.Count; j++ {
			var (
				speed float64
				err   error
			)

			if phaseType == "download" {
				speed, err = downloadSample(spec.SizeBytes)
			} else {
				var warning error
				speed, warning, err = uploadSample(spec.SizeBytes, payload)
				if warning != nil {
					throughputWarnings.Record(warning)
				}
			}
			if err != nil {
				throughputFailures.Record(err)
				continue
			}

			speeds = append(speeds, speed)
		}

		if len(speeds) > 0 {
			allSpeeds = append(allSpeeds, speeds...)
			if w != nil {
				label := sizeLabel(spec.SizeBytes)
				fmt.Fprintf(w, "    %-8s test (%d/%d): %.2f Mbps\n", label, len(speeds), spec.Count, stats.Median(speeds))
			}
		}
	}

	cancel()
	monitorWG.Wait()

	warnings := append([]string{}, throughputFailures.Message(fmt.Sprintf("%s throughput measurement failed %%d time(s); first error: %%v", phaseType))...)
	warnings = append(warnings, throughputWarnings.Message(fmt.Sprintf("%s throughput measurement used fallback timing %%d time(s); first reason: %%v", phaseType))...)
	warnings = append(warnings, loadedLatencyFailures.Message(fmt.Sprintf("%s loaded latency monitor failed %%d time(s); first error: %%v", phaseType))...)

	return PhaseMeasurement{
		Speeds:                allSpeeds,
		LoadedLatencies:       loadedLatencies,
		FailedSamples:         throughputFailures.count,
		LoadedLatencyFailures: loadedLatencyFailures.count,
		Warnings:              warnings,
	}, nil
}

func sizeLabel(size int) string {
	if size < 1_000_000 {
		return fmt.Sprintf("%dkB", size/1000)
	}
	return fmt.Sprintf("%dMB", size/1_000_000)
}

// MeasurePacketLoss sends 1000 concurrent lightweight requests and returns the loss percentage,
// number of successes, and total number of requests.
func MeasurePacketLoss() (lossPercent float64, received int, total int) {
	const totalRequests = 1000
	const concurrency = 50

	var success, failed int
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, concurrency)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			req, err := http.NewRequest("GET", baseURL+"/__down?bytes=0", nil)
			if err != nil {
				mu.Lock()
				failed++
				mu.Unlock()
				return
			}

			resp, err := defaultClient.Do(req)
			if err != nil {
				mu.Lock()
				failed++
				mu.Unlock()
				return
			}

			ok := false
			if resp.StatusCode == http.StatusOK {
				if _, err := io.Copy(io.Discard, resp.Body); err == nil {
					ok = true
				}
			}
			resp.Body.Close()

			mu.Lock()
			if ok {
				success++
			} else {
				failed++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	loss := (float64(failed) / float64(totalRequests)) * 100
	return math.Round(loss*10) / 10, success, totalRequests
}
