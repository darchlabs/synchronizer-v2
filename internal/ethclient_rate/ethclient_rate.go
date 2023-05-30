package ethclientrate

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

type BaseEthClient interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
}

type Client struct {
	ethClient BaseEthClient
	rl        *rate.Limiter

	maxRetry int
}

type Options struct {
	MaxRetry        int
	MaxRequest      int
	WindowInSeconds int
}

func NewClient(opts *Options, client BaseEthClient) *Client {
	rl := rate.NewLimiter(rate.Limit(opts.MaxRequest), opts.WindowInSeconds)

	return &Client{
		ethClient: client,
		maxRetry:  opts.MaxRetry,
		rl:        rl,
	}
}

func (cl *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (_ *big.Int, err error) {
	for i := 0; i < cl.maxRetry; i++ {
		balance, err := cl.makeBalanceAt(ctx, account, blockNumber)
		if err != nil {
			log.Printf("ethclient: attempt [%d] Client.BalanceAt cl.makeBalanceAt error %s", i, err.Error())
			continue
		}

		return balance, nil
	}

	return nil, err
}

func (cl *Client) makeBalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (_ *big.Int, err error) {
	// This is a blocking call
	err = cl.rl.Wait(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "ethclient: Client.makeBalanceAt cl.rl.Wait error")
	}

	balance, err := cl.ethClient.BalanceAt(ctx, account, blockNumber)
	if err != nil {
		return nil, errors.Wrap(err, "ethclient: Client.makeBalanceAt cl.ethClient.BalanceAt error")
	}

	return balance, nil
}
