package cloudflare

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type errReadCloser struct {
	err error
}

func (r errReadCloser) Read(_ []byte) (int, error) {
	return 0, r.err
}

func (r errReadCloser) Close() error {
	return nil
}

func useTestHTTPClient(t *testing.T, transport http.RoundTripper) {
	t.Helper()

	originalClient := defaultClient
	originalBaseURL := baseURL
	defaultClient = &http.Client{Transport: transport}
	baseURL = "https://example.test"

	t.Cleanup(func() {
		defaultClient = originalClient
		baseURL = originalBaseURL
	})
}

func TestParseServerTiming(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    float64
		wantErr string
	}{
		{
			name:   "valid value",
			header: "cfRequestDuration;dur=12.5",
			want:   12.5,
		},
		{
			name:    "missing header",
			header:  "",
			wantErr: "missing server-timing header",
		},
		{
			name:    "invalid number",
			header:  "cfRequestDuration;dur=oops",
			wantErr: "parse server timing",
		},
		{
			name:    "missing cfRequestDuration segment",
			header:  "cache;desc=HIT",
			wantErr: "cfRequestDuration not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseServerTiming(tt.header)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ParseServerTiming() error = %v", err)
				}
				if got != tt.want {
					t.Fatalf("ParseServerTiming() = %v, want %v", got, tt.want)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ParseServerTiming() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestMakeRequestReturnsDrainError(t *testing.T) {
	useTestHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Server-Timing": []string{"cfRequestDuration;dur=5"}},
			Body:       errReadCloser{err: errors.New("boom")},
			Request:    req,
		}, nil
	}))

	_, err := MakeRequest(http.MethodGet, "/__down?bytes=0", nil)
	if err == nil || !strings.Contains(err.Error(), "drain response body") {
		t.Fatalf("MakeRequest() error = %v, want drain response body error", err)
	}
}

func TestMeasurePhaseRejectsInvalidSpecs(t *testing.T) {
	tests := []struct {
		name      string
		phaseType string
		specs     []PhaseSpec
		wantErr   string
	}{
		{
			name:      "empty specs",
			phaseType: "download",
			specs:     nil,
			wantErr:   "at least one phase spec is required",
		},
		{
			name:      "invalid size",
			phaseType: "download",
			specs:     []PhaseSpec{{SizeBytes: 0, Count: 1}},
			wantErr:   "phase spec size must be positive",
		},
		{
			name:      "invalid count",
			phaseType: "upload",
			specs:     []PhaseSpec{{SizeBytes: 1000, Count: 0}},
			wantErr:   "phase spec count must be positive",
		},
		{
			name:      "unknown phase",
			phaseType: "bogus",
			specs:     []PhaseSpec{{SizeBytes: 1000, Count: 1}},
			wantErr:   "unsupported phase type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MeasurePhase(tt.phaseType, tt.specs, io.Discard)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("MeasurePhase() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestMeasurePhaseFallsBackWhenUploadServerTimingIsInvalid(t *testing.T) {
	var uploadCalls int
	useTestHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		header := http.Header{}
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/__up":
			uploadCalls++
			if uploadCalls == 1 {
				header.Set("Server-Timing", "cfRequestDuration;dur=10")
			} else {
				header.Set("Server-Timing", "cfRequestDuration;dur=oops")
			}
		case req.Method == http.MethodGet && req.URL.Path == "/__down":
			header.Set("Server-Timing", "cfRequestDuration;dur=1")
		default:
			return nil, errors.New("unexpected request")
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     header,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Request:    req,
		}, nil
	}))

	result, err := MeasurePhase("upload", []PhaseSpec{{SizeBytes: 1000, Count: 2}}, io.Discard)
	if err != nil {
		t.Fatalf("MeasurePhase() error = %v", err)
	}
	if len(result.Speeds) != 2 {
		t.Fatalf("len(result.Speeds) = %d, want 2", len(result.Speeds))
	}
	if result.FailedSamples != 0 {
		t.Fatalf("result.FailedSamples = %d, want 0", result.FailedSamples)
	}
	if len(result.Warnings) == 0 || !strings.Contains(result.Warnings[0], "upload throughput measurement used fallback timing 1 time") {
		t.Fatalf("result.Warnings = %v, want upload fallback warning", result.Warnings)
	}
}
