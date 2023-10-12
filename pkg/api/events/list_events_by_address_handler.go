package events

import (
	"encoding/json"

	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type getEventsByAddressV2Handler struct{}

type getEventsByAddressV2HandlerRequest struct {
	UserID     string
	Address    string
	Pagination *pagination.Pagination
}

type getEventsByAddressV2HandlerResponse struct {
	Events     []*EventRes                `json:"events"`
	Pagination *pagination.PaginationMeta `json:"pagination,omitempty"`
}

func (h *getEventsByAddressV2Handler) Invoke(ctx *api.Context, c *fiber.Ctx) (interface{}, int, error) {
	// define request
	req := &getEventsByAddressV2HandlerRequest{}

	// get pagination
	p := &pagination.Pagination{}
	err := p.GetPaginationFromFiber(c)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventsByAddressV2Handler.Invoke p.GetPaginationFromFiber error",
		)
	}
	req.Pagination = p

	// get user id
	req.UserID, err = api.GetUserIDFromRequestCtx(c)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventsByAddressV2Handler.Invoke c.api.GetUserIDFromRequestCtx error",
		)
	}

	// get address from params
	req.Address = c.Params("address")
	if req.Address == "" {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventsByAddressV2Handler.Invoke c.Params(address) error",
		)
	}

	return h.invoke(ctx, req)
}

// BUSINESS LOGIC
func (h *getEventsByAddressV2Handler) invoke(ctx *api.Context, req *getEventsByAddressV2HandlerRequest) (interface{}, int, error) {
	output, err := ctx.SyncEngine.SelectEventsAndABI(&sync.SelectEventsAndABIInput{
		SmartContractAddress: req.Address,
		Pagination:           req.Pagination,
	})
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventsByAddressV2Handler.invoke syncEngine.SelectEventsAndABI error",
		)
	}

	// define response
	res := &getEventsByAddressV2HandlerResponse{
		Events: make([]*EventRes, 0),
	}

	// iterate over events
	for _, event := range output.Events {
		eventRes := &EventRes{
			ID:                   event.ID,
			AbiID:                event.AbiID,
			Network:              string(event.Network),
			Name:                 event.Name,
			NodeURL:              event.NodeURL,
			Address:              event.Address,
			LatestBlockNumber:    event.LatestBlockNumber,
			SmartContractAddress: event.SmartContractAddress,
			Status:               string(event.Status),
			Error:                event.Error,
			CreatedAt:            event.CreatedAt,
			UpdatedAt:            event.UpdatedAt,
		}

		// marshal ABI record
		b, err := event.ABI.MarshalJson()
		if err != nil {
			return nil, fiber.StatusInternalServerError, errors.Wrap(
				err,
				"events: getEventsByAddressV2Handler.invoke json.Unmarshal error",
			)
		}

		// unmarshal ABI
		var abi AbiRes
		err = json.Unmarshal(b, &abi)
		if err != nil {
			return nil, fiber.StatusInternalServerError, errors.Wrap(
				err,
				"events: getEventsByAddressV2Handler.invoke json.Unmarshal error",
			)
		}
		eventRes.ABI = &abi

		res.Events = append(res.Events, eventRes)
	}

	// define pagination
	pagination := req.Pagination.GetPaginationMeta(output.TotalElements)
	res.Pagination = &pagination

	// prepare response
	return res, fiber.StatusOK, nil
}
