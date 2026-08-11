package reporter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hsblabs/universal-speedtest-cli/internal/reporter"
)

func TestPrintJSONIncludesWarningsAndNulls(t *testing.T) {
	download := 123.4
	packetLoss := 0.5

	result := reporter.Result{
		DownloadMbps: &download,
		PacketLoss:   &packetLoss,
		Received:     99,
		Total:        100,
		Warnings:     []string{"upload measurement failed 1 time(s); first error: boom"},
	}

	var buf bytes.Buffer
	if err := reporter.PrintJSON(&buf, result, true); err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"\"download_mbps\": 123.4",
		"\"upload_mbps\": null",
		"\"packet_loss_percent\": 0.5",
		"\"warnings\": [",
		"upload measurement failed 1 time(s); first error: boom",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("PrintJSON() output missing %q\n%s", want, output)
		}
	}
}

func TestPrintHumanShowsWarningsAndMissingValues(t *testing.T) {
	packetLoss := 1.2

	result := reporter.Result{
		PacketLoss: &packetLoss,
		Received:   98,
		Total:      100,
		Warnings:   []string{"network metadata unavailable: timeout"},
	}

	var buf bytes.Buffer
	reporter.PrintHuman(&buf, result)

	output := buf.String()
	for _, want := range []string{
		"N/A (insufficient data)",
		"Server Location: N/A",
		"Warnings:",
		"network metadata unavailable: timeout",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("PrintHuman() output missing %q\n%s", want, output)
		}
	}
}

func TestPrintHTMLRendersResultAndEscapesContent(t *testing.T) {
	download := 225.14
	upload := 102.87
	packetLoss := 0.1
	network := "Example <Network>"

	result := reporter.Result{
		DownloadMbps: &download,
		UploadMbps:   &upload,
		PacketLoss:   &packetLoss,
		NetworkASOrg: &network,
		Warnings:     []string{"measurement <script>alert(1)</script>"},
	}

	var buf bytes.Buffer
	if err := reporter.PrintHTML(&buf, result); err != nil {
		t.Fatalf("PrintHTML() error = %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"<!doctype html>",
		"225.14",
		"102.87",
		"Example &lt;Network&gt;",
		"measurement &lt;script&gt;alert(1)&lt;/script&gt;",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("PrintHTML() output missing %q\n%s", want, output)
		}
	}
	if strings.Contains(output, "<script>alert(1)</script>") {
		t.Fatalf("PrintHTML() output contains unescaped warning\n%s", output)
	}
	for _, forbidden := range []string{"<script", " src=", " href="} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("PrintHTML() output contains external or executable content %q\n%s", forbidden, output)
		}
	}
}
