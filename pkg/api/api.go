package api

import (
	"errors"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/env"
	"github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/internal/txsengine"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
)

type IDGenerator func() string
type DateGenerator func() time.Time

type Context struct {
	ScStorage    synchronizer.SmartContractStorage
	EventStorage synchronizer.EventStorage
	Cronjob      synchronizer.Cronjob
	TxsEngine    txsengine.TxsEngine
	Clients      *map[string]*ethclient.Client

	// Engine
	SyncEngine sync.SyncEngine

	Env     *env.Env
	IDGen   IDGenerator
	DateGen DateGenerator
}

type Handler func(*Context, *fiber.Ctx) (interface{}, int, error)

func HandleFunc(ctx *Context, fn Handler) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")
		data, statusCode, err := fn(ctx, c)
		if err != nil {
			return c.Status(statusCode).JSON(struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			})
		}

		return c.Status(statusCode).JSON(data)
	}
}

func GetUserIDFromRequestCtx(c *fiber.Ctx) (string, error) {
	id := c.Locals("user_id")
	userID, ok := id.(string)
	if !ok {
		return "", errors.New("unrecognized id type")
	}

	return userID, nil
}
