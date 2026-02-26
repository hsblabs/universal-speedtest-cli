package cloudflare

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://speed.cloudflare.com"

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

	io.Copy(io.Discard, resp.Body)
	pd.Ended = time.Now()

	for _, part := range strings.Split(resp.Header.Get("server-timing"), ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "cfRequestDuration;dur=") {
			valStr := strings.TrimPrefix(part, "cfRequestDuration;dur=")
			pd.ServerTiming, _ = strconv.ParseFloat(valStr, 64)
			break
		}
	}

	return pd, nil
}
