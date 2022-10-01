package blockchain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func NewAuth(client *ethclient.Client, pk string, value *int64, gasPrice *int64, gasLimit *uint64) (*bind.TransactOpts, error) {
	// check if pk param is defined
	if pk == "" {
		return nil, errors.New("invalid pk param")
	}
	
	// parse private key to ECDSA
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}

	// prepare public key to ECDSA
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	// get the nonce
	fromAddress :=  crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	// set value param (in wei)
	var v *big.Int
	if value != nil {
		v = big.NewInt(*value)
	} else {
		big.NewInt(0)
	}

	// set gas price param 
	var gp *big.Int
	if gasPrice != nil {
		gp = big.NewInt(*gasPrice)
	} else {
		// get recommended gas price
		gp, err = client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, err
		}
	}

	// set gas limit param (in units)
	var gl uint64
	if gasLimit != nil {
		gl = *gasLimit
	} else {
		gl = uint64(300000)
	}

	// prepare auth 
	// TODO(ca): see how to use NewKeyedTransactorWithChainID method
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(int64(5)))
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = v
	auth.GasLimit = gl
	auth.GasPrice = gp

	return auth, nil
}