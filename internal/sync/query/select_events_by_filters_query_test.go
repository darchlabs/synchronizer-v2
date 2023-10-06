package query

import (
	"testing"

	"github.com/darchlabs/synchronizer-v2/internal/test"
	"github.com/jaekwon/testify/require"
	"github.com/jmoiron/sqlx"
)

func Test_EventQuerier_SelectEventsQuery_Integration(t *testing.T) {
	test.GetTxCall(t, func(tx *sqlx.Tx, _ interface{}) {
		// Arrange
		// create basics for an event existance
		// 1. create SC
		scAddress := "0x00001"
		smartContractID := "sc-id"

		_, err := tx.Exec(`
			INSERT INTO smartcontracts (id, network, address, last_tx_block_synced, initial_block_number, created_at)
			VALUES ($1, 'testnet', $2, 0, 0, now());`,
			smartContractID,
			scAddress,
		)
		require.NoError(t, err)

		// 2. create ABI
		_, err = tx.Exec(`
			INSERT INTO abi (id, sc_address, name, type, anonymous, inputs)
			VALUES
			('abi-id-1', $1, 'sc name 1', 'type1', false, '[]'::jsonb),
			('abi-id-2', $1, 'sc name 2', 'event1', false, '[]'::jsonb),
			('abi-id-3', $1, 'sc name 3', 'event2', false, '[]'::jsonb),
			('abi-id-4', $1, 'sc name 4', 'event3', false, '[]'::jsonb),
			('abi-id-5', $1, 'sc name 5', 'type2', false, '[]'::jsonb);`,
			scAddress,
		)

		require.NoError(t, err)

		// 3. create EVENT
		_, err = tx.Exec(`
			INSERT INTO event (id, abi_id, network, name, node_url, address, latest_block_number, sc_address, status, created_at)
			VALUES
			('event-id-1', 'abi-id-1', 'testnet', 'event1', 'http://some.url', $1, 0, $1, 'synching', now()),
			('event-id-2', 'abi-id-2', 'testnet', 'event2', 'http://some.url', $1, 0, $1, 'synching', now()),
			('event-id-3', 'abi-id-3', 'testnet', 'event3', 'http://some.url', $1, 0, $1, 'running', now());`,
			scAddress,
		)
		require.NoError(t, err)

		// Act
		eq := &EventQuerier{}
		events, err := eq.SelectEventsQuery(tx, &SelectEventsQueryFilters{
			Status: "synching",
		})

		require.NoError(t, err)
		require.Equal(t, len(events), 2)

		// Assert
	})
}
