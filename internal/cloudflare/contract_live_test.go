//go:build cloudflare_live

package cloudflare

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestCloudflareLiveContract(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("download", func(t *testing.T) {
		const wantBytes = 1024
		resp, err := client.Get(baseURL + "/__down?bytes=" + strconv.Itoa(wantBytes))
		if err != nil {
			t.Fatalf("GET /__down: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /__down status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read /__down body: %v", err)
		}
		if len(body) != wantBytes {
			t.Fatalf("GET /__down body length = %d, want %d", len(body), wantBytes)
		}
		if resp.Header.Get("Server-Timing") == "" {
			t.Error("GET /__down missing Server-Timing header")
		}
	})

	t.Run("upload", func(t *testing.T) {
		const wantBytes = 1024
		payload := bytes.Repeat([]byte("0"), wantBytes)
		req, err := http.NewRequest(http.MethodPost, baseURL+"/__up", bytes.NewReader(payload))
		if err != nil {
			t.Fatalf("create POST /__up request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /__up: %v", err)
		}
		defer resp.Body.Close()
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			t.Fatalf("read /__up body: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST /__up status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		gotBytes, err := strconv.Atoi(resp.Header.Get("cf-meta-upload-bytes"))
		if err != nil {
			t.Fatalf("cf-meta-upload-bytes = %q: %v", resp.Header.Get("cf-meta-upload-bytes"), err)
		}
		if gotBytes != wantBytes {
			t.Fatalf("cf-meta-upload-bytes = %d, want %d", gotBytes, wantBytes)
		}
		if resp.Header.Get("Server-Timing") == "" {
			t.Error("POST /__up missing Server-Timing header")
		}
	})

	t.Run("metadata", func(t *testing.T) {
		meta, err := FetchMeta()
		if err != nil {
			t.Fatalf("GET /meta: %v", err)
		}
		if meta.ASN <= 0 || meta.ClientIP == "" || meta.Colo.City == "" {
			t.Fatalf("/meta missing required fields: asn=%d ip=%t city=%q", meta.ASN, meta.ClientIP != "", meta.Colo.City)
		}
	})

	t.Logf("validated Cloudflare contract at %s", baseURL)
}
