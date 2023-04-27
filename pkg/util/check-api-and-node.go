package util

import (
	"context"
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/ethereum/go-ethereum/ethclient"
)

func CheckAndGetApis(contract *smartcontract.SmartContract, networksApiUrls map[string]string, networksApiKeys map[string]string) (string, string, error) {
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

func CheckAndGetNodeURL(contract *smartcontract.SmartContract, networksNodeUrl map[string]string) (string, error) {
	// Get contract network
	contractNewtork := string(contract.Network)

	// Check if the contract node url is valid
	err := NodeURLIsValid(contract.NodeURL, contractNewtork)
	if err != nil {
		// If it is not valid, check the backend node url (obtained from the env)
		networkNodeUrl := networksNodeUrl[contractNewtork]
		err := NodeURLIsValid(networkNodeUrl, contractNewtork)
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

func NodeURLIsValid(nodeUrl string, network string) error {
	if network == "" {
		return fmt.Errorf("\nthe network was not provided for checking the node url")
	}

	if nodeUrl == "" {
		return fmt.Errorf("\nthe node url is empty")
	}

	networkChainId := SupportedNetworks[network]
	if networkChainId == 0 {
		return fmt.Errorf("\nthe network %s is not currently supported by darchlabs", network)
	}

	client, err := ethclient.Dial(nodeUrl)
	if err != nil {
		return err
	}

	clientNetworkId, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	if clientNetworkId.Int64() != networkChainId {
		return fmt.Errorf("/nthe node url network %v chain id doesn't much the given network %d chain id", clientNetworkId, clientNetworkId)
	}

	return nil
}

var SupportedNetworks = map[string]int64{
	"ethereum": 1,
	"polygon":  137,
	"celo":     42220,
}
