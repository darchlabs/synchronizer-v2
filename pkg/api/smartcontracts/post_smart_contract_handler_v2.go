package smartcontracts

import (
	"encoding/json"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type postSmartContractV2Handler struct {
	validate *validator.Validate
}

type postSmartContractV2HandlerRequest struct {
	SmartContract *SmartContractReq `json:"smartcontract"`
}

// HTTP SERVER LOGIC
func (h *postSmartContractV2Handler) Invoke(ctx *api.Context, c *fiber.Ctx) (interface{}, interface{}, int, error) {
	var req postSmartContractV2HandlerRequest
	err := c.BodyParser(&req)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: postSmartContractV2Handler.Invoke c.BodyParser error",
		)
	}

	err = h.validate.Struct(req.SmartContract)
	if err != nil {
		return nil, nil, fiber.StatusBadRequest, errors.Wrap(
			err,
			"smartcontracts: postSmartContractV2HandlerRequest.Invoke h.validate.Struct error",
		)
	}

	req.SmartContract.UserID, err = api.GetUserIDFromRequestCtx(c)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: postSmartContractV2Handler.Invoke c.api.GetUserIDFromRequestCtx error",
		)
	}

	return h.invoke(ctx, &req)
}

// BUSINESS LOGIC
func (h *postSmartContractV2Handler) invoke(ctx *api.Context, req *postSmartContractV2HandlerRequest) (interface{}, interface{}, int, error) {
	// get and validate node url
	nodeURL := req.SmartContract.NodeURL
	network := string(req.SmartContract.Network)
	err := util.NodeURLIsValid(nodeURL, network)
	if err != nil {
		networksEtherscanURL, err := util.ParseStringifiedMap(ctx.Env.NetworksNodeURL)
		if err != nil {
			// CHECKPOINT
			return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
				err,
				"smartcontracts: postSmartContractV2Handler.invoke util.ParseStringifiedMap can't valid ethclient error",
			)
		}

		nodeURL = networksEtherscanURL[network]
	}

	abi := make([]*storage.ABIRecord, 0)

	// Loop over ABI
	for _, a := range req.SmartContract.ABI {
		// input
		bytes, err := json.Marshal(a.Inputs)
		if err != nil {
			return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
				err,
				"smartcontracts: postSmartContractV2Handler.invoke json.Marshal input abi error",
			)

		}

		ipts := make([]*storage.InputABI, 0)
		err = json.Unmarshal(bytes, &ipts)
		if err != nil {
			return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
				err,
				"smartcontracts: postSmartContractV2Handler.invoke json.Unmarshal input abi error",
			)

		}

		// create ABI
		abi = append(abi, &storage.ABIRecord{
			SmartContractAddress: req.SmartContract.Address,
			Name:                 a.Name,
			Type:                 a.Type,
			Anonymous:            a.Anonymous,
			Inputs:               ipts,
		})
	}

	output, err := ctx.SyncEngine.InsertAtomicSmartContract(&sync.InsertAtomicSmartContractInput{
		UserID:     req.SmartContract.UserID,
		Name:       req.SmartContract.Name,
		NodeURL:    nodeURL,
		WebhookURL: req.SmartContract.WebhookURL,
		SmartContract: &storage.SmartContractRecord{
			Address: req.SmartContract.Address,
			Network: storage.Network(req.SmartContract.Network),
		},
		ABI: abi,
	})
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: postSmartContractV2Handler.invoke syncEngine.InsertAtomicSmartContract  error",
		)
	}

	abiRes := make([]*AbiReq, 0)
	for _, a := range output.ABI {
		inputReq, err := TransformInputsJsonToArray(a.Inputs)
		if err != nil {
			return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
				err,
				"smartcontracts: postSmartContractV2Handler.invoke transformInputsJsonToArray error",
			)
		}

		abiRes = append(abiRes, &AbiReq{
			Name:      a.Name,
			Type:      a.Type,
			Anonymous: a.Anonymous,
			Inputs:    inputReq,
		})
	}
	scRes := &SmartContractReq{
		Network:    string(output.SmartContract.Network),
		Name:       output.SmartContractUser.Name,
		Address:    output.SmartContract.Address,
		NodeURL:    output.SmartContractUser.NodeURL,
		WebhookURL: output.SmartContractUser.WebhookURL,
		ABI:        abiRes,
	}

	return scRes, nil, fiber.StatusCreated, nil
}
