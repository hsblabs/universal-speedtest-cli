---
schema_version: "2026-08-15"
okf_version: "0.2"
type: guide
title: unispeedtest — command-line internet speed test
description: A command-line network benchmark that measures download and upload throughput, latency, jitter, and packet loss over Cloudflare speed test endpoints.
resource: https://github.com/hsblabs/universal-speedtest-cli
tags:
  - cli
  - go
  - network
  - speedtest
status: stable
hsblabs:
  sidebar:
    label: Overview
    order: 1
---

`unispeedtest` measures the quality of a network connection from the terminal. It currently drives Cloudflare speed test endpoints, and the provider layer is structured so other backends can be added later.

## What it measures

| Metric | How it is derived |
| --- | --- |
| Download throughput | 90th percentile of sampled Mbps values |
| Upload throughput | 90th percentile of sampled Mbps values |
| Unloaded latency | Median of 20 samples |
| Loaded latency | Measured separately during the download and upload phases |
| Jitter | Mean absolute difference between consecutive unloaded latency samples |
| Packet loss | 1000 requests at concurrency 50 |
| Network metadata | Cloudflare colo, ASN and AS organization, public IP |

The reports also grade the connection for streaming, gaming, and video chat. The full sampling plan and the grading thresholds are in [How measurements work](./how-it-works.md).

## Install

### From GitHub Releases

The installer script resolves the latest release, verifies the archive against the release `checksums.txt` with SHA-256, and installs to `/usr/local/bin`:

```sh
curl -fsSL https://raw.githubusercontent.com/hsblabs/universal-speedtest-cli/main/install.sh | sh
```

Set `INSTALL_DIR` to install elsewhere:

```sh
curl -fsSL https://raw.githubusercontent.com/hsblabs/universal-speedtest-cli/main/install.sh | INSTALL_DIR="$HOME/.local/bin" sh
```

To install by hand, download an archive from [Releases](https://github.com/hsblabs/universal-speedtest-cli/releases) and place the `unispeedtest` binary somewhere on your `PATH`.

### With Go

```sh
go install github.com/hsblabs/universal-speedtest-cli/cmd/unispeedtest@latest
```

The binary is named `unispeedtest`.

## Run

```sh
unispeedtest
```

A full run takes roughly a minute and prints a colored summary. Set `NO_COLOR=1` to disable ANSI colors.

## Options

| Flag | Effect |
| --- | --- |
| `-html <path>` | Write a self-contained HTML report to `<path>` |
| `-html-title <title>` | Append a title to the HTML report; requires `-html` |
| `-json` | Print compact single-line JSON |
| `-pretty` | Print indented JSON; implies `-json` |
| `-v`, `--version` | Print the CLI version and exit |

```sh
unispeedtest -html report.html -html-title "Home Wi-Fi"
unispeedtest -json
unispeedtest -pretty
unispeedtest --version
```

Passing `-html-title` without `-html` is an error and exits with status 2.

Both report formats are covered in [Output formats](./output.md).

Under `-json`, progress output is suppressed so stdout stays parseable.
