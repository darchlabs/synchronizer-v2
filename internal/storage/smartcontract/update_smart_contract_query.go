package smartcontractstorage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
)

func (s *Storage) UpdateSmartContract(sc *smartcontract.SmartContract) (*smartcontract.SmartContract, error) {
	// get current sc
	current, err := s.GetSmartContractByAddress(sc.Address)
	if err != nil {
		return nil, fmt.Errorf("smartcontract not found with address=%s", sc.Address)
	}

	// prepare dinamic sql
	query := "UPDATE smartcontracts SET "
	args := []interface{}{}

	// check if name is changed
	if sc.Name != "" {
		query += "name = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.Name)
	}

	// check if nodeurl is changed
	if sc.NodeURL != "" {
		query += "node_url = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.NodeURL)
	}

	// check if webhook is changed
	query += "webhook = $" + strconv.Itoa(len(args)+1) + ", "
	args = append(args, sc.Webhook)

	// add where sql condition
	query = strings.TrimSuffix(query, ", ")
	query += " WHERE id = $" + strconv.Itoa(len(args)+1)
	args = append(args, current.ID)

	// execute query
	_, err = s.storage.DB.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	// get updated smartcontract
	updatedSmartContract, err := s.GetSmartContractById(current.ID)
	if err != nil {
		return nil, err
	}

	return updatedSmartContract, nil
}
