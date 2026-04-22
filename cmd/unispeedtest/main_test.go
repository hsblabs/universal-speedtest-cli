package main

import (
	"bytes"
	"io"
	"runtime/debug"
	"strings"
	"testing"
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
	benchmarkMain = func(jsonOut, prettyOut bool, stdout, stderr io.Writer) int {
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
	benchmarkMain = func(jsonOut, prettyOut bool, stdout, stderr io.Writer) int {
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

	benchmarkMain = func(jsonOut, prettyOut bool, stdout, stderr io.Writer) int {
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
