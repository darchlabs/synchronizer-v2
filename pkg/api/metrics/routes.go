package metrics

import (
	"github.com/darchlabs/synchronizer-v2"
	txsengine "github.com/darchlabs/synchronizer-v2/internal/txsengine"
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	SmartContractStorage synchronizer.SmartContractStorage
	TransactionStorage   synchronizer.TransactionStorage
	EventStorage         synchronizer.EventStorage
	Engine               txsengine.TxsEngine
}

func Route(app *fiber.App, ctx Context) {
	// Engine status related endpoints
	app.Get("/api/v1/metrics/status", getEngineStatus(ctx))
	app.Post("/api/v1/metrics/start", startEngine(ctx))
	app.Post("/api/v1/metrics/stop", stopEngine(ctx))

	// Transactions data related endpoints
	app.Get("/api/v1/metrics/transactions", listTransactions(ctx))
	app.Get("/api/v1/metrics/transactions/:address", listSmartContractTransactions(ctx))
	app.Get("/api/v1/metrics/transactions/:address/failed", listSmartContractFailedTransactions(ctx))
	app.Get("/api/v1/metrics/addresses/:address", listSmartContractActiveAddresses(ctx))
	app.Get("/api/v1/metrics/tvl/:address/current", getSmartContractCurrentTVL(ctx))
	app.Get("/api/v1/metrics/tvl/:address", listSmartContractTVLs(ctx))
	app.Get("/api/v1/metrics/gas/:address", listSmartContractGasSpent(ctx))
	app.Get("/api/v1/metrics/gas/:address/total", getSmartContractTotalGasSpent(ctx))
	app.Get("/api/v1/metrics/value/:address/total", getSmartContractTotalValueTransferred(ctx))
}
