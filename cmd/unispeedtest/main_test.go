package main

import (
    "runtime/debug"
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
}
