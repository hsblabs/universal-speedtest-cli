package cloudflare

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/hsblabs/universal-speedtest-cli/internal/stats"
)

// MeasureSpeed converts a transfer of byteCount bytes completed in durationMs milliseconds
// into megabits per second.
func MeasureSpeed(byteCount int, durationMs float64) float64 {
	return (float64(byteCount*8) / (durationMs / 1000)) / 1e6
}

// MeasureLatency sends 20 lightweight requests and returns the unloaded latency samples in ms.
func MeasureLatency() []float64 {
	var measurements []float64
	for i := 0; i < 20; i++ {
		pd, err := MakeRequest("GET", "/__down?bytes=0", nil)
		if err == nil && !pd.TTFB.IsZero() {
			dur := pd.TTFB.Sub(pd.Started).Seconds()*1000 - pd.ServerTiming
			if dur > 0 {
				measurements = append(measurements, dur)
			}
		}
	}
	return measurements
}

// backgroundLatencyMonitor periodically samples latency while other measurements run.
func backgroundLatencyMonitor(ctx context.Context, results *[]float64, mu *sync.Mutex) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pd, err := MakeRequest("GET", "/__down?bytes=0", nil)
			if err == nil && !pd.TTFB.IsZero() {
				dur := pd.TTFB.Sub(pd.Started).Seconds()*1000 - pd.ServerTiming
				if dur > 0 {
					mu.Lock()
					*results = append(*results, dur)
					mu.Unlock()
				}
			}
		}
	}
}

// MeasurePhase runs either a download or upload measurement phase across the given sizes and
// repetition counts. Progress is written to w if non-nil.
// Returns all sampled speeds (Mbps) and the loaded latency samples collected during the phase.
func MeasurePhase(phaseType string, sizes []int, counts []int, w io.Writer) ([]float64, []float64) {
	var allSpeeds []float64
	var loadedLatencies []float64
	var mu sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go backgroundLatencyMonitor(ctx, &loadedLatencies, &mu)

	for i, size := range sizes {
		count := counts[i]
		var speeds []float64

		var payload []byte
		if phaseType != "download" {
			payload = bytes.Repeat([]byte("0"), size)
		}
		for j := 0; j < count; j++ {
			var pd PerfData
			var err error

			if phaseType == "download" {
				pd, err = MakeRequest("GET", fmt.Sprintf("/__down?bytes=%d", size), nil)
				if err == nil && !pd.TTFB.IsZero() {
					transferTime := pd.Ended.Sub(pd.TTFB).Seconds() * 1000
					speeds = append(speeds, MeasureSpeed(size, transferTime))
				}
			} else {
				pd, err = MakeRequest("POST", "/__up", payload)
				if err == nil && pd.ServerTiming > 0 {
					speeds = append(speeds, MeasureSpeed(size, pd.ServerTiming))
				}
			}
		}

		if len(speeds) > 0 {
			allSpeeds = append(allSpeeds, speeds...)
			if w != nil {
				label := sizeLabel(size)
				fmt.Fprintf(w, "    %-8s test (%d/%d): %.2f Mbps\n", label, len(speeds), count, stats.Median(speeds))
			}
		}
	}

	return allSpeeds, loadedLatencies
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

			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				if resp.StatusCode == 200 {
					io.Copy(io.Discard, resp.Body)
					success++
				} else {
					failed++
				}
				resp.Body.Close()
			} else {
				failed++
			}
		}()
	}
	wg.Wait()

	loss := (float64(failed) / float64(totalRequests)) * 100
	return math.Round(loss*10) / 10, success, totalRequests
}
