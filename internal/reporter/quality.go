package reporter

// QualityScore holds the pass/fail result for each use-case category.
type QualityScore struct {
	Streaming bool
	Gaming    bool
	Chatting  bool
}

// EvaluateQuality determines connection quality for common use cases based on the given metrics.
// download and upload are in Mbps, latency and jitterMs in milliseconds, packetLoss in percent.
func EvaluateQuality(download, upload, latency, jitterMs, packetLoss float64) QualityScore {
	return QualityScore{
		Streaming: download > 5.0 && latency < 100 && packetLoss < 2.0,
		Gaming:    latency < 50 && jitterMs < 20 && packetLoss < 1.0,
		Chatting:  download > 2.0 && upload > 2.0 && latency < 100 && jitterMs < 30 && packetLoss < 1.0,
	}
}
