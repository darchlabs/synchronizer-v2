package txsengine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (t *T) getTransactionsFromEtherscan(apiURL string, apiKey string, address string, startBlock int64, lastBlock int64) ([]*transaction.Transaction, error) {
	var txs []*transaction.Transaction

	type Response struct {
		Status  string                     `json:"status"`
		Message string                     `json:"message"`
		Result  []*transaction.Transaction `json:"result"`
	}

	// parse url
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing API URL: %v", err)
	}

	// define params and encode
	params := url.Values{}
	params.Set("module", "account")
	params.Set("action", "txlist")
	params.Set("address", address)
	params.Set("startblock", strconv.Itoa(int(startBlock)))
	params.Set("endblock", strconv.Itoa(int(lastBlock)))
	params.Set("sort", "asc")
	params.Set("apikey", apiKey)
	u.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	res, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", res.StatusCode)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var body Response
	err = json.Unmarshal(b, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	if body.Status != "1" {
		return nil, fmt.Errorf("API request failed with status: %s, message: %s", body.Status, body.Message)
	}

	txs = body.Result
	return txs, err
}
