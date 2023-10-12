package smartcontracts

import (
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type listSmartContractV2Handler struct{}

type listSmartContractV2HandlerRequest struct {
	UserID     string
	Pagination *pagination.Pagination
}

// type listSmartContractV2HandlerResponse struct {
// 	SmartContracts []*SmartContractRequest `json:"smartcontracts"`
// }

func (h *listSmartContractV2Handler) Invoke(ctx *api.Context, c *fiber.Ctx) (interface{}, interface{}, int, error) {
	// define request
	req := &listSmartContractV2HandlerRequest{}

	// get pagination
	p := &pagination.Pagination{}
	err := p.GetPaginationFromFiber(c)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: postSmartContractV2Handler.Invoke p.GetPaginationFromFiber error",
		)
	}
	req.Pagination = p

	req.UserID, err = api.GetUserIDFromRequestCtx(c)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: postSmartContractV2Handler.Invoke c.api.GetUserIDFromRequestCtx error",
		)
	}

	return h.invoke(ctx, req)
}

// BUSINESS LOGIC
func (h *listSmartContractV2Handler) invoke(ctx *api.Context, req *listSmartContractV2HandlerRequest) (interface{}, interface{}, int, error) {
	output, err := ctx.SyncEngine.SelectUserSmartContractsWithEvents(&sync.SelectUserSmartContractsWithEventsInput{
		UserID:     req.UserID,
		Pagination: req.Pagination,
	})
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: postSmartContractV2Handler.invoke syncEngine.InsertAtomicSmartContract  error",
		)
	}

	// parse smart contracts
	contractsRes := make([]*SmartContractResponse, 0)
	for _, sc := range output.SmartContracts {
		scRes := &SmartContractResponse{
			ID:                 sc.ID,
			Network:            string(sc.Network),
			Name:               sc.Name,
			Address:            sc.Address,
			WebhookURL:         sc.WebhookURL,
			Status:             sc.Status,
			LastTxBlockSynced:  sc.LastTxBlockSynced,
			InitialBlockNumber: sc.InitialBlockNumber,
			Error:              sc.Error,
		}

		// parse events
		events := make([]*EventResponse, 0)
		for _, e := range sc.Events {
			eventRes := &EventResponse{
				ID:     e.ID,
				Status: e.Status,
				Error:  e.Error,
			}
			events = append(events, eventRes)
		}

		scRes.Events = events
		contractsRes = append(contractsRes, scRes)
	}

	// define meta response
	meta := make(map[string]interface{})
	meta["pagination"] = req.Pagination.GetPaginationMeta(output.TotalElements)

	// prepare response
	return contractsRes, meta, fiber.StatusOK, nil
}
