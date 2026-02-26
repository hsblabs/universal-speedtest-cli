package cloudflare

import "time"

// CloudFlareMeta represents metadata returned by the Cloudflare speed test API.
type CloudFlareMeta struct {
	ASN            int    `json:"asn"`
	ASOrganization string `json:"asOrganization"`
	ClientIP       string `json:"clientIp"`
	City           string `json:"city"`
	Colo           struct {
		IATA   string  `json:"iata"`
		Lat    float64 `json:"lat"`
		Lon    float64 `json:"lon"`
		CCA2   string  `json:"cca2"`
		Region string  `json:"region"`
		City   string  `json:"city"`
	} `json:"colo"`
}

// PerfData holds timing data captured during an HTTP request.
type PerfData struct {
	Started      time.Time
	TTFB         time.Time
	Ended        time.Time
	ServerTiming float64
}
