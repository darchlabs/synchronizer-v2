package events

import (
	"encoding/json"

	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type getEventDataV2Handler struct{}

type getEventDataV2HandlerRequest struct {
	UserID     string
	Address    string
	EventName  string
	Pagination *pagination.Pagination
}

type getEventDataV2HandlerResponse struct {
	Datas      []*EventDataRes            `json:"datas"`
	Event      *EventRes                  `json:"event,omitempty"`
	Pagination *pagination.PaginationMeta `json:"pagination,omitempty"`
}

func (h *getEventDataV2Handler) Invoke(ctx *api.Context, c *fiber.Ctx) (interface{}, int, error) {
	// define request
	req := &getEventDataV2HandlerRequest{}

	// get pagination
	p := &pagination.Pagination{}
	err := p.GetPaginationFromFiber(c)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventDataV2Handler.Invoke p.GetPaginationFromFiber error",
		)
	}
	req.Pagination = p

	// get user id
	req.UserID, err = api.GetUserIDFromRequestCtx(c)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventDataV2Handler.Invoke c.api.GetUserIDFromRequestCtx error",
		)
	}

	// get address from params
	req.Address = c.Params("address")
	if req.Address == "" {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventDataV2Handler.Invoke c.Params(address) error",
		)
	}

	// get eventName from params
	req.EventName = c.Params("event_name")
	if req.Address == "" {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventDataV2Handler.Invoke c.Params(event_name) error",
		)
	}

	return h.invoke(ctx, req)
}

// BUSINESS LOGIC
func (h *getEventDataV2Handler) invoke(ctx *api.Context, req *getEventDataV2HandlerRequest) (interface{}, int, error) {
	// define response and pagination
	res := &getEventDataV2HandlerResponse{
		Datas: make([]*EventDataRes, 0),
	}

	output, err := ctx.SyncEngine.SelectEventData(&sync.SelectEventDataInput{
		SmartContractAddress: req.Address,
		EventName:            req.EventName,
		Pagination:           req.Pagination,
	})
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"events: getEventDataV2Handler.invoke syncEngine.SelectEventData error",
		)
	}

	// prepare pagination
	pagination := req.Pagination.GetPaginationMeta(output.TotalElements)
	res.Pagination = &pagination

	// prepare event
	if output.Event != nil {
		res.Event = &EventRes{
			ID:                   output.Event.ID,
			AbiID:                output.Event.AbiID,
			Network:              string(output.Event.Network),
			Name:                 output.Event.Name,
			NodeURL:              output.Event.NodeURL,
			Address:              output.Event.Address,
			LatestBlockNumber:    output.Event.LatestBlockNumber,
			SmartContractAddress: output.Event.SmartContractAddress,
			Status:               string(output.Event.Status),
			Error:                output.Event.Error,
			CreatedAt:            output.Event.CreatedAt,
			UpdatedAt:            output.Event.UpdatedAt,
		}

		// prepare ABI if exists
		if output.Event.ABI != nil {
			// marshal ABI record
			b, err := output.Event.ABI.MarshalJson()
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
			res.Event.ABI = &abi

		}

		// iterate over events
		for _, data := range output.EventsData {
			dataReq := &EventDataRes{
				ID:          data.ID,
				EventID:     data.EventID,
				Tx:          data.Tx,
				BlockNumber: data.BlockNumber,
				Data:        data.Data,
				CreatedAt:   data.CreatedAt,
			}

			res.Datas = append(res.Datas, dataReq)
		}
	}

	// prepare response
	return res, fiber.StatusOK, nil
}
