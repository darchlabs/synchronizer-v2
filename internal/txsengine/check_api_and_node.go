package txsengine

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
)

func checkAndGetApis(contract *smartcontract.SmartContract, networksApiUrls map[string]string, networksApiKeys map[string]string) (string, string, error) {
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

func checkAndGetNodeURL(contract *smartcontract.SmartContract, networksNodeUrl map[string]string) (string, error) {
	// Get contract network
	contractNewtork := string(contract.Network)

	// Check if the contract node url is valid
	err := util.NodeURLIsValid(contract.NodeURL, contractNewtork)
	if err != nil {
		// If it is not valid, check the backend node url (obtained from the env)
		networkNodeUrl := networksNodeUrl[contractNewtork]
		err := util.NodeURLIsValid(networkNodeUrl, contractNewtork)
		if err != nil {
			// If the backend node url is not valid, return err since both nodes are bad
			return "", err
		}

		// return the backend node url if it is valid
		return networkNodeUrl, nil

	}

	// return the contract node url if it is valid
	return contract.NodeURL, nil
}
