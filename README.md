# unispeedtest

[日本語](docs/README/ja.md)

`unispeedtest` is a command-line network benchmark. It currently uses Cloudflare speed test endpoints and is designed to be extensible to other providers.

## Features

- Download throughput (90th percentile of sampled Mbps values)
- Upload throughput (90th percentile of sampled Mbps values)
- Unloaded latency (median of 20 samples)
- Loaded latency during download and upload phases
- Jitter (average absolute difference between consecutive unloaded latency samples)
- Packet loss (1000 requests, concurrency 50)
- Network metadata (Cloudflare colo, ASN/AS organization, public IP)

## Installation

### 1) Install from GitHub Releases (recommended)

Download the latest archive from [Releases](https://github.com/hsblabs/universal-speedtest-cli/releases), then place `unispeedtest` in your `PATH`.

Or use the installer script:

```sh
curl -fsSL https://raw.githubusercontent.com/hsblabs/universal-speedtest-cli/main/install.sh | sh
```

By default, it installs to `/usr/local/bin` (`INSTALL_DIR` can be overridden).
The installer verifies SHA-256 checksums using the release `checksums.txt`.

### 2) Install with Go

```sh
go install github.com/hsblabs/universal-speedtest-cli/cmd/unispeedtest@latest
```

This installs the binary as `unispeedtest`.

## Usage

```sh
unispeedtest
```

Options:

- `-json`: output compact JSON
- `-pretty`: output indented JSON (implies `-json`)

Examples:

```sh
unispeedtest -json
unispeedtest -pretty
```

## JSON output shape

On partial failures, affected metrics are emitted as `null` and a `warnings` array is added so callers can distinguish missing data from real zero values.

```json
{
  "download_mbps": 225.14,
  "upload_mbps": 102.87,
  "latency_ms": {
    "unloaded": 12.41,
    "loaded_down": 35.09,
    "loaded_up": 41.22,
    "jitter": 1.98
  },
  "packet_loss_percent": 0.1,
  "server_colo": "Tokyo",
  "network_asn": "AS2516",
  "network_as_org": "KDDI CORPORATION",
  "ip": "203.0.113.10",
  "warnings": [
    "upload loaded latency unavailable: no samples collected"
  ]
}
```

## Development

Run tests:

```sh
go test ./...
```

Build:

```sh
go build -trimpath -ldflags="-s -w" -o dist/unispeedtest ./cmd/unispeedtest
```

Tips:

- Set `NO_COLOR=1` to disable ANSI color output.

## Releasing

- Follow Conventional Commits in merged PRs (`fix:`, `feat:`, `feat!:` / `BREAKING CHANGE:`).
- Pushes to `main` update or create a release PR via `release-please`.
- Merging that release PR updates `version.txt` and `CHANGELOG.md`.
- The release workflow then creates the corresponding `vX.Y.Z` tag and publishes artifacts with GoReleaser.

## License

MIT
