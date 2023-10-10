package sync

import (
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/pkg/errors"
)

type SelectUserSmartContractsWithEventsInput struct {
	UserID     string
	Pagination *pagination.Pagination
}

type SelectUserSmartContractsWithEventsOutput struct {
	SmartContracts []*query.UserSmartContractOutput
	TotalElements  int64
}

func (ng *Engine) SelectUserSmartContractsWithEvents(input *SelectUserSmartContractsWithEventsInput) (*SelectUserSmartContractsWithEventsOutput, error) {
	// Select user smart contracts
	smartContracts, err := ng.SmartContractQuerier.SelectUserSmartContractsQuery(ng.database, input.UserID, input.Pagination)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectUserSmartContractsWithEvents ng.SmartContractQuerier.SelectUserSmartContractsQuery error")
	}

	// Get smart contract addresses
	addresses := make([]string, 0)

	constractsMap := make(map[string]*query.UserSmartContractOutput)

	for _, sc := range smartContracts {
		constractsMap[sc.Address] = sc
		constractsMap[sc.Address].Events = make([]storage.EventRecord, 0)
		addresses = append(addresses, sc.Address)
	}

	// Select events by addresses
	events, err := ng.EventQuerier.SelectEventsByAddressesListQuery(ng.database, addresses)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectUserSmartContractsWithEvents ng.EventQuerier.SelectEventsByAddressesListQuery error")
	}

	// Add events to smart contracts
	for _, event := range events {
		contract, ok := constractsMap[event.Address]
		if !ok {
			continue
		}
		contract.Events = append(contract.Events, *event)
	}

	// Get the total count of user's smart contracts
	totalElements, err := ng.SmartContractQuerier.SelectCountUserSmartContractsQuery(ng.database, input.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectUserSmartContractsWithEvents ng.SmartContractQuerier.CountUserSmartContractsQuery error")
	}

	return &SelectUserSmartContractsWithEventsOutput{
		SmartContracts: smartContracts,
		TotalElements:  totalElements,
	}, nil
}
