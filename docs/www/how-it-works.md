---
schema_version: "2026-08-15"
okf_version: "0.2"
type: reference
title: How unispeedtest measures your connection
description: The sampling plan behind each metric, why throughput uses the 90th percentile, the network quality score thresholds, and how partial failures are handled.
resource: https://github.com/hsblabs/universal-speedtest-cli
tags:
  - methodology
  - network
  - speedtest
status: stable
hsblabs:
  sidebar:
    label: How it works
    order: 3
---

Speed test numbers only mean something if you know how they were produced. This page documents the sampling plan and the statistics `unispeedtest` applies, so a result can be compared against other tools — or against itself over time — without guessing.

All measurements run against Cloudflare speed test endpoints over a client with a 30 second request timeout.

## Latency

20 samples are taken with a zero-byte `GET` before any load is applied.

Each sample measures the time from request start to the first response byte, then **subtracts Cloudflare's `Server-Timing` value** so the result reflects network round-trip time rather than server processing. When the header is missing or unparseable, the raw client timing is used instead and a warning names how many samples fell back.

- **Unloaded latency** — median of the samples.
- **Jitter** — mean absolute difference between consecutive samples. Requires at least 2 samples.
- **Loaded latency** — median of the latency samples collected *during* the download and upload phases, reported separately for each. This is the number that reflects bufferbloat.

## Throughput

Each phase transfers a fixed ladder of payload sizes, small to large, so both fast and slow connections land on a usable set of samples.

| Payload | Download runs | Upload runs |
| --- | --- | --- |
| 101 KB | 10 | 8 |
| 1 MB | 8 | 6 |
| 10 MB | 6 | 4 |
| 25 MB | 4 | 4 |

The two directions time different windows, because the byte movement they care about happens at different points in the exchange:

- **Download** — first response byte to last response byte. Connection setup and time-to-first-byte are excluded, so the figure is transfer rate rather than end-to-end request time.
- **Upload** — request start until the request body has been fully written. If that timing is unavailable, the full end-to-end duration is used instead and a warning is emitted; such samples read slower than the link actually is.

The reported value is the **90th percentile** of those samples, with linear interpolation between neighbours. A high percentile rather than the mean or max: the mean is dragged down by the small payloads that never reach line rate, and the max would report a single lucky burst.

## Packet loss

1000 requests at concurrency 50. The result is the percentage that did not come back, and the terminal report also shows the raw received-over-total count.

## Network quality score

The terminal and HTML reports grade the connection **Good** or **Poor** for three use cases. Every condition must hold for a Good grade:

| Use case | Conditions |
| --- | --- |
| Video streaming | download > 5 Mbps, unloaded latency < 100 ms, packet loss < 2% |
| Online gaming | unloaded latency < 50 ms, jitter < 20 ms, packet loss < 1% |
| Video chatting | download > 2 Mbps, upload > 2 Mbps, unloaded latency < 100 ms, jitter < 30 ms, packet loss < 1% |

If any of the five inputs is missing, the whole score reads `N/A (insufficient data)` rather than grading on partial evidence.

The score is derived entirely from the metrics above, and is **not** part of the JSON output — a consumer that needs it applies these thresholds itself.

## Failures

The run distinguishes two kinds of failure.

**Fatal** — a phase produced no usable samples at all. Latency, download, upload, and packet loss each abort the run this way. Nothing is reported.

**Partial** — some samples failed, or a derived metric could not be computed. The run continues, the affected field is `null` in JSON, and a warning explains why. Individual sample failures are collapsed into one warning per category carrying the count and the first error.

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Run completed, or `--version` / `-h` printed and exited |
| `1` | A measurement phase failed fatally |
| `2` | Invalid flags, including `-html-title` without `-html` |

A `0` exit does not imply a complete result. Check `warnings` — see [Output formats](./output.md).

## Progress output

Progress lines are written to stdout during a normal run and suppressed by `-json`, so piping to a JSON parser needs no extra redirection.
