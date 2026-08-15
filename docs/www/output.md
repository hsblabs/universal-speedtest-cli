---
schema_version: "2026-08-15"
okf_version: "0.2"
type: reference
title: JSON and HTML output
description: The JSON payload unispeedtest emits, its nullable fields and warnings, and the self-contained HTML speed test report.
resource: https://github.com/hsblabs/universal-speedtest-cli
tags:
  - cli
  - json
  - reporting
status: stable
hsblabs:
  sidebar:
    label: Output formats
    order: 2
---

`unispeedtest` writes a human-readable summary to stdout by default. `-json` and `-html` add machine-readable and shareable formats; they compose, so a single run can produce all three.

## JSON

`-json` prints one line; `-pretty` prints the same document indented.

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

Progress output is suppressed under `-json`, so stdout is a clean JSON document.

### Partial failures

A run can succeed overall while an individual measurement fails. When that happens the affected field is emitted as `null` and an explanation is appended to `warnings`, so a consumer can tell missing data apart from a genuine `0`.

Two things to handle when parsing:

- **Every metric field is nullable.** `download_mbps`, `upload_mbps`, all four `latency_ms` members, `packet_loss_percent`, and the network metadata strings can all be `null`.
- **`warnings` is omitted entirely when empty** — it is not an empty array. Read it as "absent or non-empty".

A zero exit code means the run finished, not that every metric is present. See [How measurements work](./how-it-works.md) for which failures are fatal and which degrade to a warning.

The network quality score shown in the terminal and HTML reports is not included in the JSON. Its thresholds are documented in [How measurements work](./how-it-works.md#network-quality-score) if you need to reproduce it.

## HTML report

`-html <path>` writes a responsive single-file report with no external assets — no CDN, no network access needed to view it. An existing file at the path is overwritten.

The measurement time is stored as Unix epoch milliseconds and rendered by inline JavaScript in the viewer's own locale and time zone, so a report shared across regions reads correctly for each reader.

`-html-title` sets a suffix on both the document title and the page heading:

```sh
unispeedtest -html report.html -html-title "Home Wi-Fi"
```

produces `Internet Speed Report - Home Wi-Fi`.
