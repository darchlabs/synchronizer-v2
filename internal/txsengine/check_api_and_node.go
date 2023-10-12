package txsengine

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
)

func checkAndGetApis(contract *query.SelectSmartContractQueryOutput, networksApiUrls map[string]string, networksApiKeys map[string]string) (string, string, error) {
	contractNewtork := string(contract.Network)
	etherscanApiURL := networksApiUrls[contractNewtork]
	etherscanApiKey := networksApiKeys[contractNewtork]

	if etherscanApiURL == "" {
		return "", "", fmt.Errorf("\nempty etherscan api url for the %s network", contractNewtork)
	}

	if etherscanApiKey == "" {
		return "", "", fmt.Errorf("\nempty etherscan api key for the %s network", contractNewtork)
	}

	return etherscanApiURL, etherscanApiKey, nil
}
