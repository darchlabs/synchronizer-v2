package metrics

import (
	"github.com/darchlabs/synchronizer-v2"
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	SmartContractStorage synchronizer.SmartContractStorage
	TransactionStorage   synchronizer.TransactionStorage
	EventStorage         synchronizer.EventStorage
}

func Route(app *fiber.App, ctx Context) {
	app.Get("/api/v1/metrics/transactions", listTransactions(ctx))
	app.Get("/api/v1/metrics/transactions/:address", listSmartContractTransactions(ctx))
	app.Get("/api/v1/metrics/transactions/:address/total", listSmartContractTotalTransactions(ctx))
	app.Get("/api/v1/metrics/transactions/:address/failed", listSmartContractFailedTransactions(ctx))
	app.Get("/api/v1/metrics/transactions/:address/failed/total", listSmartContractTotalFailedTransactions(ctx))

	app.Get("/api/v1/metrics/addresses/:address", listSmartContractActiveAddresses(ctx))
	app.Get("/api/v1/metrics/addresses/:address/total", listSmartContractTotalActiveAddresses(ctx))

	app.Get("/api/v1/metrics/tvl/:address/current", listSmartContractCurrentTVL(ctx))
	app.Get("/api/v1/metrics/tvl/:address", listSmartContractTVLs(ctx))

	app.Get("/api/v1/metrics/gas/:address", listSmartContractGasSpent(ctx))
	app.Get("/api/v1/metrics/gas/:address/total", listSmartContractTotalGasSpent(ctx))

	app.Get("/api/v1/metrics/value/:address/total", listSmartContractTotalValueTransferred(ctx))

	// TODO(nb): Create this handler when the logs part is refactorized
	// app.Get("/api/v1/metrics/events/:address/total", listSmartContractTotalEvents(ctx))
}
