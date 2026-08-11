package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/hsblabs/universal-speedtest-cli/internal/reporter"
)

func TestHasVersionFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no flags", args: []string{}, want: false},
		{name: "json only", args: []string{"-json"}, want: false},
		{name: "short version", args: []string{"-v"}, want: true},
		{name: "long version", args: []string{"--version"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasVersionFlag(tt.args); got != tt.want {
				t.Fatalf("hasVersionFlag(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestResolvedVersion(t *testing.T) {
	oldVersion, oldReader := version, buildInfoReader
	t.Cleanup(func() {
		version = oldVersion
		buildInfoReader = oldReader
	})

	t.Run("prefers injected version", func(t *testing.T) {
		version = "v1.2.3"
		buildInfoReader = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: "v0.9.0"}}, true
		}

		if got := resolvedVersion(); got != "v1.2.3" {
			t.Fatalf("resolvedVersion() = %q, want %q", got, "v1.2.3")
		}
	})

	t.Run("uses build info version", func(t *testing.T) {
		version = ""
		buildInfoReader = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: "v0.9.0"}}, true
		}

		if got := resolvedVersion(); got != "v0.9.0" {
			t.Fatalf("resolvedVersion() = %q, want %q", got, "v0.9.0")
		}
	})

	t.Run("falls back to dev", func(t *testing.T) {
		version = ""
		buildInfoReader = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
		}

		if got := resolvedVersion(); got != "dev" {
			t.Fatalf("resolvedVersion() = %q, want %q", got, "dev")
		}
	})

	t.Run("nil build info falls back to dev", func(t *testing.T) {
		version = ""
		buildInfoReader = func() (*debug.BuildInfo, bool) {
			return nil, true
		}

		if got := resolvedVersion(); got != "dev" {
			t.Fatalf("resolvedVersion() = %q, want %q", got, "dev")
		}
	})
}

func TestRunVersionFlagPrintsAndExits(t *testing.T) {
	oldVersion, oldBenchmark := version, benchmarkMain
	t.Cleanup(func() {
		version = oldVersion
		benchmarkMain = oldBenchmark
	})

	version = "v1.2.3"
	benchmarkMain = func(jsonOut, prettyOut bool, htmlPath, htmlTitle string, stdout, stderr io.Writer) int {
		t.Fatal("benchmark path should not run for version flags")
		return 1
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := run([]string{"--version"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run(--version) exit code = %d, want 0", exitCode)
	}

	if got := stdout.String(); got != "v1.2.3\n" {
		t.Fatalf("stdout = %q, want %q", got, "v1.2.3\n")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunShortVersionFlagPrintsAndExits(t *testing.T) {
	oldVersion, oldBenchmark := version, benchmarkMain
	t.Cleanup(func() {
		version = oldVersion
		benchmarkMain = oldBenchmark
	})

	version = "v9.9.9"
	benchmarkMain = func(jsonOut, prettyOut bool, htmlPath, htmlTitle string, stdout, stderr io.Writer) int {
		t.Fatal("benchmark path should not run for version flags")
		return 1
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := run([]string{"-v"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run(-v) exit code = %d, want 0", exitCode)
	}

	if got := stdout.String(); got != "v9.9.9\n" {
		t.Fatalf("stdout = %q, want %q", got, "v9.9.9\n")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunHelpFlagPrintsUsageAndExitsZero(t *testing.T) {
	oldBenchmark := benchmarkMain
	t.Cleanup(func() {
		benchmarkMain = oldBenchmark
	})

	benchmarkMain = func(jsonOut, prettyOut bool, htmlPath, htmlTitle string, stdout, stderr io.Writer) int {
		t.Fatal("benchmark path should not run for help flags")
		return 1
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := run([]string{"-h"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run(-h) exit code = %d, want 0", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	if got := stderr.String(); !strings.Contains(got, "Usage of unispeedtest:") {
		t.Fatalf("stderr = %q, want usage output", got)
	}
}

func TestRunHTMLFlagsPassOutputOptions(t *testing.T) {
	oldBenchmark := benchmarkMain
	t.Cleanup(func() {
		benchmarkMain = oldBenchmark
	})

	benchmarkMain = func(jsonOut, prettyOut bool, htmlPath, htmlTitle string, stdout, stderr io.Writer) int {
		if htmlPath != "report.html" {
			t.Fatalf("htmlPath = %q, want %q", htmlPath, "report.html")
		}
		if htmlTitle != "Home" {
			t.Fatalf("htmlTitle = %q, want %q", htmlTitle, "Home")
		}
		return 0
	}

	if exitCode := run([]string{"-html", "report.html", "-html-title", "Home"}, io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}
}

func TestRunRejectsHTMLTitleWithoutHTMLPath(t *testing.T) {
	oldBenchmark := benchmarkMain
	t.Cleanup(func() {
		benchmarkMain = oldBenchmark
	})

	benchmarkMain = func(jsonOut, prettyOut bool, htmlPath, htmlTitle string, stdout, stderr io.Writer) int {
		t.Fatal("benchmark path should not run for invalid flags")
		return 1
	}

	var stderr bytes.Buffer
	if exitCode := run([]string{"-html-title", "Home"}, io.Discard, &stderr); exitCode != 2 {
		t.Fatalf("run() exit code = %d, want 2", exitCode)
	}
	if got := stderr.String(); !strings.Contains(got, "-html-title requires -html") {
		t.Fatalf("stderr = %q, want dependency error", got)
	}
}

func TestWriteHTMLReport(t *testing.T) {
	download := 123.4
	path := filepath.Join(t.TempDir(), "report.html")
	if err := os.WriteFile(path, []byte("old report"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if err := writeHTMLReport(path, "Home", reporter.Result{DownloadMbps: &download}); err != nil {
		t.Fatalf("writeHTMLReport() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if output := string(data); !strings.Contains(output, "123.40") {
		t.Fatalf("HTML report missing download speed:\n%s", output)
	}
	if output := string(data); !strings.Contains(output, "Internet Speed Report - Home") {
		t.Fatalf("HTML report missing custom title:\n%s", output)
	}
	if strings.Contains(string(data), "old report") {
		t.Fatalf("HTML report did not replace existing file:\n%s", data)
	}
}
