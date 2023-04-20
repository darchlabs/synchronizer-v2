package transaction

import (
	"github.com/darchlabs/synchronizer-v2"
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	TransactionStorage synchronizer.TransactionStorage
}

func Route(app *fiber.App, ctx Context) {
	// app.Get("/api/v1/transactions", listTransactionsHandler(ctx))
	// app.Get("/api/v1/transactions/count", getTransactionsCount(ctx))
	// app.Get("/api/v1/transactions/wallets/active", getActiveWallets(ctx))
	// app.Get("/api/v1/transactions/wallets/active/count", getActiveWalletsCount(ctx))
	// app.Get("/api/v1/transactions/:address/tvl", getSmartContractTVL(ctx))
	// app.Get("/api/v1/transactions/:address/gas/total-amount", getTotalGasPaid(ctx))
	// app.Get("/api/v1/transactions/:address/gas/total-price", getTotalGasPrice(ctx))

}
