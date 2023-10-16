package txsengine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Response struct {
	Result  string `json:"result"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (t *T) GetCurrentBlockNumberFromEtherscan(apiURL, apiKey string) (int64, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// parse
	u, err := url.Parse(apiURL)
	if err != nil {
		return 0, fmt.Errorf("error parsing API URL: %v", err)
	}

	// define params and encode
	params := url.Values{}
	params.Set("module", "block")
	params.Set("action", "getblocknobytime")
	params.Set("timestamp", timestamp)
	params.Set("closest", "before")
	params.Set("apikey", apiKey)
	u.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	res, err := t.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP request failed with status code: %d", res.StatusCode)
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP request failed with status code: %d", res.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var body Response
	err = json.Unmarshal(b, &body)
	if err != nil {
		return 0, fmt.Errorf("failed to decode response body: %v", err)
	}

	if body.Status != "1" || body.Message != "OK" {
		return 0, fmt.Errorf("API request failed with status: %s, message: %s", body.Status, body.Message)
	}

	blockNumber, err := strconv.ParseInt(body.Result, 10, 64)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}
