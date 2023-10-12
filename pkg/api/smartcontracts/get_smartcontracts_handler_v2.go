package smartcontracts

import (
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type getSmartContractV2Handler struct{}

type getSmartContractV2HandlerRequest struct {
	UserID     string
	Pagination *pagination.Pagination
}

type getSmartContractV2HandlerResponse struct {
	SmartContracts []*SmartContractResponse   `json:"contracts"`
	Pagination     *pagination.PaginationMeta `json:"pagination,omitempty"`
}

func (h *getSmartContractV2Handler) Invoke(ctx *api.Context, c *fiber.Ctx) (interface{}, int, error) {
	// define request
	req := &getSmartContractV2HandlerRequest{}

	// get pagination
	p := &pagination.Pagination{}
	err := p.GetPaginationFromFiber(c)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: getSmartContractV2Handler.Invoke p.GetPaginationFromFiber error",
		)
	}
	req.Pagination = p

	// get user id
	req.UserID, err = api.GetUserIDFromRequestCtx(c)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: getSmartContractV2Handler.Invoke c.api.GetUserIDFromRequestCtx error",
		)
	}

	return h.invoke(ctx, req)
}

// BUSINESS LOGIC
func (h *getSmartContractV2Handler) invoke(ctx *api.Context, req *getSmartContractV2HandlerRequest) (interface{}, int, error) {
	output, err := ctx.SyncEngine.SelectUserSmartContractsWithEvents(&sync.SelectUserSmartContractsWithEventsInput{
		UserID:     req.UserID,
		Pagination: req.Pagination,
	})
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: getSmartContractV2Handler.invoke syncEngine.InsertAtomicSmartContract error",
		)
	}

	// define response
	res := &getSmartContractV2HandlerResponse{
		SmartContracts: make([]*SmartContractResponse, 0),
	}

	// parse smart contracts
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
				Name:   e.Name,
				Status: e.Status,
				Error:  e.Error,
			}
			events = append(events, eventRes)
		}

		scRes.Events = events
		res.SmartContracts = append(res.SmartContracts, scRes)
	}

	// define pagination
	pagination := req.Pagination.GetPaginationMeta(output.TotalElements)
	res.Pagination = &pagination

	// prepare response
	return res, fiber.StatusOK, nil
}
