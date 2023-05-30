package util

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

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
