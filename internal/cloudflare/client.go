package cloudflare

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"
)

var baseURL = "https://speed.cloudflare.com"

var defaultClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		DisableKeepAlives:   true,
	},
	Timeout: 30 * time.Second,
}

// MakeRequest performs an HTTP request to the Cloudflare speed test endpoint and
// returns timing data. The path should begin with "/".
func MakeRequest(method, path string, payload []byte) (PerfData, error) {
	var pd PerfData

	req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return pd, err
	}

	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() { pd.TTFB = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	pd.Started = time.Now()
	resp, err := defaultClient.Do(req)
	if err != nil {
		return pd, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pd, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return pd, fmt.Errorf("drain response body: %w", err)
	}
	pd.Ended = time.Now()
	pd.ServerTimingHeader = resp.Header.Get("server-timing")

	return pd, nil
}

// ParseServerTiming extracts the cfRequestDuration value from a Server-Timing header.
func ParseServerTiming(header string) (float64, error) {
	if header == "" {
		return 0, errors.New("missing server-timing header")
	}

	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "cfRequestDuration;dur=") {
			valStr := strings.TrimPrefix(part, "cfRequestDuration;dur=")
			value, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return 0, fmt.Errorf("parse server timing %q: %w", valStr, err)
			}
			return value, nil
		}
	}

	return 0, fmt.Errorf("cfRequestDuration not found in server-timing header %q", header)
}
