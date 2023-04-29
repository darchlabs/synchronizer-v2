package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
)

func GetTransactionsFromEtherscan(apiUrl string, apiKey string, address string, startBlock int64) ([]*transaction.Transaction, error) {
	var txs []*transaction.Transaction

	type Response struct {
		Result []*transaction.Transaction `json"result"`
	}

	var res Response

	url := fmt.Sprintf("%s?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", apiUrl, address, fmt.Sprint(startBlock), apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	txs = res.Result
	return txs, err
}
