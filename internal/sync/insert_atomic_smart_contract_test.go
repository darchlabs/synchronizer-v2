package sync

import (
	"testing"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/darchlabs/synchronizer-v2/internal/test"
	uuid "github.com/google/uuid"
	"github.com/jaekwon/testify/require"
	"github.com/jmoiron/sqlx"
)

func getEngineForTest() (*Engine, error) {
	engine := &Engine{
		ABIQuerier:               query.NewABIQuerier(nil, uuid.NewString, time.Now),
		SmartContractQuerier:     query.NewSmartContractQuerier(nil, uuid.NewString, time.Now),
		SmartContractUserQuerier: query.NewSmartContractUserQuerier(nil, uuid.NewString, time.Now),
		InputQuerier:             query.NewInputQuerier(nil, uuid.NewString, time.Now),
		EventQuerier:             query.NewEventsQuerier(nil, uuid.NewString, time.Now),

		dateGen: time.Now,
		idGen:   uuid.NewString,
		logger:  nil,
	}

	return engine, nil
}

func Test_InsertAtomicSmartContract_FirstInsertion(t *testing.T) {
	test.GetDBCall(t, func(db *sqlx.DB, _ interface{}) {
		ng, err := getEngineForTest()

		require.NoError(t, err)
		require.NotNil(t, ng)

		ng.database = test.GetTestDB(db)

		// MAKING INSERT ATOMIC SC TEST
		input := &InsertAtomicSmartContractInput{
			UserID:     "user-id",
			Name:       "test-contract-id",
			WebhookURL: "https://nodeurl.com",
			NodeURL:    "https://nodeurl.com",

			SmartContract: &storage.SmartContractRecord{
				Network:           "test-net",
				Address:           "0x0000000000000000000000000000000000000001",
				LastTxBlockSynced: 1212121,
				CreatedAt:         time.Now(),
			},
			ABI: []*storage.ABIRecord{
				{
					Name:      "abi-record-id",
					Type:      "type1",
					Anonymous: true,
					Inputs: []map[string]interface{}{
						{"indexed": false, "internalType": "it", "name": "foo", "type": "foo"},
					},
				},
				{
					Name:      "abi-record-id-2",
					Type:      "event",
					Anonymous: false,
					Inputs: []map[string]interface{}{
						{"indexed": false, "internalType": "it", "name": "foo", "type": "foo"},
					},
				},
			},
		}
		out, err := ng.InsertAtomicSmartContract(input)

		require.NoError(t, err)
		require.NotNil(t, out)

	})
}

func Test_InsertAtomicSmartContract_NthInsertion(t *testing.T) {
	test.GetDBCall(t, func(db *sqlx.DB, _ interface{}) {
		ng, err := getEngineForTest()

		require.NoError(t, err)
		require.NotNil(t, ng)

		ng.database = test.GetTestDB(db)

		// MAKING INSERT ATOMIC SC TEST
		input := &InsertAtomicSmartContractInput{
			UserID:     "user-id",
			Name:       "test-contract-id",
			WebhookURL: "https://nodeurl.com",
			NodeURL:    "https://nodeurl.com",

			SmartContract: &storage.SmartContractRecord{
				Network:           "test-net",
				Address:           "0x0000000000000000000000000000000000000001",
				LastTxBlockSynced: 1212121,
				CreatedAt:         time.Now(),
			},
			ABI: []*storage.ABIRecord{
				{
					Name:      "abi-record-id",
					Type:      "type1",
					Anonymous: true,
					Inputs: []map[string]interface{}{
						{"indexed": false, "internalType": "it", "name": "foo", "type": "foo"},
					},
				},
				{
					Name:      "abi-record-id-2",
					Type:      "event",
					Anonymous: false,
					Inputs: []map[string]interface{}{
						{"indexed": false, "internalType": "it", "name": "foo", "type": "foo"},
					},
				},
			},
		}
		out, err := ng.InsertAtomicSmartContract(input)

		require.NoError(t, err)
		require.NotNil(t, out)

		// MAKING INSERT ATOMIC SC TEST
		input = &InsertAtomicSmartContractInput{
			UserID:     "user-id-2",
			Name:       "test-contract-id",
			WebhookURL: "https://nodeurl.com",
			NodeURL:    "https://nodeurl.com",

			SmartContract: &storage.SmartContractRecord{
				Network:           "test-net",
				Address:           "0x0000000000000000000000000000000000000001",
				LastTxBlockSynced: 1212121,
				CreatedAt:         time.Now(),
			},
			ABI: []*storage.ABIRecord{
				{
					Name:      "abi-record-id",
					Type:      "type1",
					Anonymous: true,
					Inputs: []map[string]interface{}{
						{"indexed": false, "internalType": "it", "name": "foo", "type": "foo"},
					},
				},
				{
					Name:      "abi-record-id-2",
					Type:      "event",
					Anonymous: false,
					Inputs: []map[string]interface{}{
						{"indexed": false, "internalType": "it", "name": "foo", "type": "foo"},
					},
				},
			},
		}

		out2, err := ng.InsertAtomicSmartContract(input)

		require.NoError(t, err)
		require.Equal(t, out2.SmartContractUser.SmartContractAddress, out.SmartContract.Address)
	})
}
