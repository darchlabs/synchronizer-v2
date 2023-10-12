package sync

import (
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/pkg/errors"
)

type SelectEventsAndABIInput struct {
	SmartContractAddress string
	EventStatus          string
	Pagination           *pagination.Pagination
}

type SelectEventsAndABIOutput struct {
	Events        []*storage.EventRecord
	TotalElements int64
}

func (ng *Engine) SelectEventsAndABI(input *SelectEventsAndABIInput) (*SelectEventsAndABIOutput, error) {
	// Select events by status
	events, err := ng.EventQuerier.SelectEventsQuery(ng.database, &query.SelectEventsQueryFilters{
		Status:               input.EventStatus,
		SmartContractAddress: input.SmartContractAddress,
		Pagination:           input.Pagination,
	})
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectEventsAndABI ng.eventQuerier.SelectEventsQuery error")
	}

	// Select abi from an id list
	abiIDs := make([]string, 0)
	scAddresses := make([]string, 0)
	for _, ev := range events {
		abiIDs = append(abiIDs, ev.AbiID)
		scAddresses = append(scAddresses, ev.SmartContractAddress)
	}
	abi, err := ng.ABIQuerier.SelectABIByIDs(ng.database, abiIDs)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectEventsAndABI ng.eventQuerier.SelectABIByIDs error")
	}

	// Select smart contract
	smartContracts, err := ng.SmartContractQuerier.SelectSmartContractsByAddressesList(ng.database, scAddresses)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectEventsAndABI ng.SmartContractQuerier.SelectSmartContractsByAddressesList error")
	}

	scUsers, err := ng.SmartContractUserQuerier.SmartContractUsersByIDListQuery(
		ng.database,
		scAddresses,
	)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectEventsAndABI ng.SmartContractQuerier.SmartContractUsersByIDListQuery error")
	}

	// mapping of abis and smart contracts
	abiMap := make(map[string]*storage.ABIRecord)
	for _, a := range abi {
		abiMap[a.ID] = a
	}
	scMap := make(map[string]*storage.SmartContractRecord)
	for _, sc := range smartContracts {
		scMap[sc.Address] = sc
	}
	scuMap := make(map[string][]*storage.SmartContractUserRecord)
	for _, scu := range scUsers {
		if _, ok := scuMap[scu.SmartContractAddress]; !ok {
			scuMap[scu.SmartContractAddress] = make([]*storage.SmartContractUserRecord, 0)
		}
		scuMap[scu.SmartContractAddress] = append(scuMap[scu.SmartContractAddress], scu)
	}

	// Link abi ans sc with events
	for _, ev := range events {
		ev.ABI = abiMap[ev.AbiID]
		ev.SmartContract = scMap[ev.SmartContractAddress]
		ev.SmartContractUsers = scuMap[ev.SmartContractAddress]
	}

	// Count total elements if pagination is defined
	var totalElements int64
	if input.Pagination != nil {
		totalElements, err = ng.EventQuerier.SelectCountEventsQuery(ng.database, &query.SelectEventsQueryFilters{
			Status:               input.EventStatus,
			SmartContractAddress: input.SmartContractAddress,
		})
		if err != nil {
			return nil, errors.Wrap(err, "sync: Engine.SelectEventsAndABI ng.eventQuerier.CountEventsQuery error")
		}
	}

	return &SelectEventsAndABIOutput{
		Events:        events,
		TotalElements: totalElements,
	}, nil
}
