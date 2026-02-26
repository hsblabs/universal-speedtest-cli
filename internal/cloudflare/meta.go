package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// FetchMeta retrieves network metadata (ASN, IP, colo) from the Cloudflare speed test API.
func FetchMeta() (CloudFlareMeta, error) {
	var meta CloudFlareMeta

	req, err := http.NewRequest("GET", baseURL+"/meta", nil)
	if err != nil {
		return meta, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", baseURL+"/")
	req.Header.Set("Origin", baseURL)

	resp, err := defaultClient.Do(req)
	if err != nil {
		return meta, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return meta, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&meta)
	return meta, err
}
