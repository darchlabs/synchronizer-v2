package sync

import (
	"database/sql"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// TODO(mt): Define all needed inputs for atomic inserting smart contract
type InsertAtomicSmartContractInput struct {
	UserID     string
	Name       string
	WebhookURL string
	NodeURL    string

	SmartContract *storage.SmartContractRecord
	ABI           []*storage.ABIRecord
}

type InsertAtomicSmartContractOutput struct {
	SmartContract     *storage.SmartContractRecord
	ABI               []*storage.ABIRecord
	Events            []*storage.EventRecord
	SmartContractUser *storage.SmartContractUserRecord
}

// InsertAtomicSmartContract is the function in charge of handling database logic
// for atomic inserting smart contract. To insert a smart contract exists a dependendy
// related to other tables.
// abi for sc --> [n] inputs for abi --> events per sc --> smart_contract -> smart_contract_per_user
//
// the custom logic meant to this is as follows
// Pre:
//			Validate if SmartContract is already inserted.
//			If so, insert only smartcontract_users table record
//
//1. insert abi
//2. insert n inputs (from abi)
//3. insert event ()
//4. insert smartcontract
//5. insert smart_contract_user
func (ng *Engine) InsertAtomicSmartContract(input *InsertAtomicSmartContractInput) (*InsertAtomicSmartContractOutput, error) {
	output, err := ng.checkBeforeInsertAtomicSmartcontract(input)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.InsertAtomicSmartContract ng.checkBeforeInsertAtomicSmartcontract error")
	}
	if output != nil {
		return output, nil
	}

	output, err = ng.insertAtomicSmartContract(input)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.InsertAtomicSmartContract ng.insertAtomicSmartContract error")
	}

	return output, nil
}

func (ng *Engine) checkBeforeInsertAtomicSmartcontract(input *InsertAtomicSmartContractInput) (*InsertAtomicSmartContractOutput, error) {
	// select smartcontract
	sc, err := ng.smartContractQuerier.SelectSmartContractByAddressQuery(ng.database, input.SmartContract.Address)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "ng.smartContractQuerier.SelectSmartContractByAddressQuery error")
	}

	now := ng.dateGen()

	// create smart_contract_user
	scUser := &storage.SmartContractUserRecord{
		ID:                   ng.idGen(),
		UserID:               input.UserID,
		SmartContractAddress: input.SmartContract.Address,
		WebhookURL:           input.WebhookURL,
		NodeURL:              input.NodeURL,
		Status:               storage.SmartContractStatusIdle,
		CreatedAt:            now,
		Name:                 input.Name,
	}
	err = ng.smartContractUserQuerier.UpsertSmartContractUserQuery(ng.database, scUser)
	if err != nil {
		return nil, errors.Wrap(err, "ng.smartContractUserQuerier.UpsertSmartContractUserQuery error")
	}

	//ABI               *storage.ABIRecord
	abi, err := ng.abiQuerier.SelectABIByAddressQuery(ng.database, sc.Address)
	if err != nil {
		return nil, errors.Wrap(err, "ng.ABIQuerier.SelectABIByAddressQuery error")
	}

	//inputs, err := ng.inputQuerier.SelectInputByABIIDQuery(ng.database, abi.ID)
	//if err != nil {
	//return nil, errors.Wrap(err, "ng.inputQuerier.SelectInputByABIIDQuery error")
	//}
	//abi.Inputs = inputs

	events, err := ng.eventQuerier.SelectEventsByAddressQuery(ng.database, sc.Address)
	if err != nil {
		return nil, errors.Wrap(err, "ng.eventQuerier.SelectEventsByAddressQuery error")
	}
	//Events            []*storage.EventRecord
	//SmartContractUser *storage.SmartContractUserRecord

	return &InsertAtomicSmartContractOutput{
		SmartContract:     sc,
		SmartContractUser: scUser,
		ABI:               abi,
		Events:            events,
	}, nil
}

func (ng *Engine) insertAtomicSmartContract(input *InsertAtomicSmartContractInput) (*InsertAtomicSmartContractOutput, error) {
	var output InsertAtomicSmartContractOutput
	err := ng.WithTransaction(ng.database, func(txx *sqlx.Tx) error {
		now := ng.dateGen()
		// ids
		smartContractID := ng.idGen()

		// Insert SmartContract
		input.SmartContract.ID = smartContractID
		input.SmartContract.CreatedAt = now
		err := ng.smartContractQuerier.InsertSmartContractQuery(txx, input.SmartContract)
		if err != nil {
			return errors.Wrap(err, "ng.smartContractQuerier.InsertSmartContractQuery error")
		}

		// Insert ABI and Input
		err = ng.abiQuerier.InsertABIBatchQuery(txx, input.ABI, input.SmartContract.Address)
		if err != nil {
			return errors.Wrap(err, "ng.abiQuerier.InsertABIQuery error")
		}

		// insert events
		events := make([]*storage.EventRecord, 0)
		for _, abi := range input.ABI {
			if abi.Type == "event" {
				events = append(events, &storage.EventRecord{
					Network:              storage.EventNetwork(input.SmartContract.Network),
					NodeURL:              input.NodeURL,
					LatestBlockNumber:    int64(0), // For explicity since default value por numbers is 0
					Status:               storage.EventStatusSynching,
					Address:              input.SmartContract.Address,
					SmartContractAddress: input.SmartContract.Address,
				})
			}
		}
		err = ng.eventQuerier.InsertEventBatchQuery(txx, events, input.SmartContract.Address)
		if err != nil {
			return errors.Wrap(err, "ng.eventQuerier.InsertEventBatchQuery error")
		}

		// Insert SmartContractUser
		smartContractUserInput := &storage.SmartContractUserRecord{
			ID:                   ng.idGen(),
			UserID:               input.UserID,
			SmartContractAddress: input.SmartContract.Address,
			WebhookURL:           input.WebhookURL,
			NodeURL:              input.NodeURL,
			Status:               storage.SmartContractStatusIdle,
			CreatedAt:            now,
			Name:                 input.Name,
		}
		err = ng.smartContractUserQuerier.UpsertSmartContractUserQuery(txx, smartContractUserInput)
		if err != nil {
			return errors.Wrap(err, "ng.smartContractUserQuerier.UpsertSmartContractUserQuery error")
		}
		output.SmartContractUser = smartContractUserInput
		output.ABI = input.ABI
		output.SmartContract = input.SmartContract

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "ng.WithTransaction error")
	}

	return &output, nil

}
