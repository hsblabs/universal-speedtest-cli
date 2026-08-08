package cloudflare

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"testing"
	"time"
)

func TestUploadSampleUsesRequestWrittenTiming(t *testing.T) {
	useTestHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		trace := httptrace.ContextClientTrace(req.Context())
		if trace == nil || trace.WroteRequest == nil {
			t.Fatal("request trace does not expose WroteRequest")
		}

		time.Sleep(10 * time.Millisecond)
		trace.WroteRequest(httptrace.WroteRequestInfo{})

		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Server-Timing": []string{"cfRequestDuration;dur=0.001"},
			},
			Body:    io.NopCloser(strings.NewReader("")),
			Request: req,
		}, nil
	}))

	speed, warning, err := uploadSample(10_000, bytes.Repeat([]byte("0"), 10_000))
	if err != nil {
		t.Fatalf("uploadSample() error = %v", err)
	}
	if warning != nil {
		t.Fatalf("uploadSample() warning = %v, want nil", warning)
	}
	if speed <= 0 || speed >= 1_000 {
		t.Fatalf("uploadSample() speed = %v, want a positive client-timed speed below 1000 Mbps", speed)
	}
}

func TestLatencySampleFallsBackToClientTiming(t *testing.T) {
	useTestHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		trace := httptrace.ContextClientTrace(req.Context())
		if trace == nil || trace.GotFirstResponseByte == nil {
			t.Fatal("request trace does not expose GotFirstResponseByte")
		}

		time.Sleep(10 * time.Millisecond)
		trace.GotFirstResponseByte()

		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Server-Timing": []string{"cfSpeedEdge;dur=3"},
			},
			Body:    io.NopCloser(strings.NewReader("")),
			Request: req,
		}, nil
	}))

	latency, err := latencySample()
	if err != nil {
		t.Fatalf("latencySample() error = %v", err)
	}
	if latency <= 0 {
		t.Fatalf("latencySample() = %v, want a positive client-timed latency", latency)
	}
}
