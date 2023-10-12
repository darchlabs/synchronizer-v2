package sync

import (
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/pkg/errors"
)

type SelectEventDataInput struct {
	SmartContractAddress string
	EventName            string
	Pagination           *pagination.Pagination
}

type SelectEventDataOutput struct {
	EventsData    []*storage.EventDataRecord
	Event         *storage.EventRecord
	TotalElements int64
}

func (ng *Engine) SelectEventData(input *SelectEventDataInput) (*SelectEventDataOutput, error) {
	// Select events by status
	eventsData, err := ng.EventDataQuerier.SelectEventDataQuery(ng.database, &query.SelectEventDataQueryFilters{
		SmartContractAddress: input.SmartContractAddress,
		EventName:            input.EventName,
		Pagination:           input.Pagination,
	})
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectEventData ng.eventQuerier.SelectEventDataQuery error")
	}

	// get event record
	events, err := ng.EventQuerier.SelectEventsQuery(ng.database, &query.SelectEventsQueryFilters{
		SmartContractAddress: input.SmartContractAddress,
		EventName:            input.EventName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectEventData ng.eventQuerier.SelectEventsQuery error")
	}
	if len(events) == 0 {
		err = errors.New("no event found")
		return nil, errors.Wrap(err, "sync: Engine.SelectEventData ng.eventQuerier.SelectEventsQuery error")
	}

	// Select abi from an id list
	abi, err := ng.ABIQuerier.SelectABIByIDs(ng.database, []string{events[0].AbiID})
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectEventData ng.eventQuerier.SelectABIByIDs error")
	}
	if len(abi) == 0 {
		err = errors.New("no abi found")
		return nil, errors.Wrap(err, "sync: Engine.SelectEventData ng.eventQuerier.SelectABIByIDs error")
	}
	events[0].ABI = abi[0]

	// Count total elements if pagination is defined
	var totalElements int64
	if input.Pagination != nil {
		totalElements, err = ng.EventDataQuerier.SelectCountEventDataQuery(ng.database, &query.SelectCountEventDataQueryFilters{
			SmartContractAddress: input.SmartContractAddress,
			EventName:            input.EventName,
		})
		if err != nil {
			return nil, errors.Wrap(err, "sync: Engine.SelectEventData ng.eventQuerier.SelectCountEventDataQuery error")
		}
	}

	return &SelectEventDataOutput{
		EventsData:    eventsData,
		Event:         events[0],
		TotalElements: totalElements,
	}, nil
}
