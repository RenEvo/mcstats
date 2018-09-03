package simplehttp

import (
	"bytes"
	"encoding/json"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

var (
	httpClient = retryablehttp.NewClient()
)

// GetJSON from the server
func GetJSON(url string, v interface{}) error {
	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "mcstats/1.0.0")
	resp, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}

// PostJSON to the server
func PostJSON(url string, d interface{}, v interface{}) error {
	reqBody, err := json.Marshal(d)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "mcstats/1.0.0")
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}
