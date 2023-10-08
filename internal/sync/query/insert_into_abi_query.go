package query

import (
	"encoding/json"
	"fmt"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (aq *ABIQuerier) InsertABIQuery(qCtx storage.QueryContext, input *storage.ABIRecord) error {
	var inputsjson []byte
	if len(input.Inputs) != 0 {
		bytes, err := json.Marshal(input.Inputs)
		if err != nil {
			return errors.Wrap(err, "query ABIQuerier.InsertABIQuery abi json.Marshal error")
		}
		inputsjson = bytes
	}
	fmt.Printf("~~~> %s\n", string(inputsjson))

	_, err := qCtx.Exec(`
		INSERT INTO abi (id, sc_address, name, type, anonymous, inputs)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		input.ID,
		input.SmartContractAddress,
		input.Name,
		input.Type,
		input.Anonymous,
		inputsjson,
	)
	if err != nil {
		return errors.Wrap(err, "query ABIQuerier.InsertABIQuery abi tx.Exec error")
	}

	return nil
}
