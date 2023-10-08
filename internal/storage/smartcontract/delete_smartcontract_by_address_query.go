package smartcontractstorage

import "github.com/pkg/errors"

func (s *Storage) DeleteSmartContractByAddress(address string) error {
	// list events by address from storage
	events, err := s.eventStorage.ListAllEvents()
	if err != nil {
		return errors.Wrap(err, "smartcontractstorage: Storage.DeleteSmartContractByAddress s.EventStorage.ListAllEvents error")
	}

	// delete events from storage
	for _, ev := range events {
		if ev.Address == address {
			err = s.eventStorage.DeleteEvent(address, ev.Abi.Name)
			if err != nil {
				return errors.Wrap(err, "smartcontractstorage: Storage.DeleteSmartContractByAddress s.eventStorage.DeleteEvent error")
			}
		}
	}

	// get smartcontract using the address
	sc, err := s.GetSmartContractByAddress(address)
	if err != nil {
		return errors.Wrap(err, "smartcontractstorage: Storage.DeleteSmartContractByAddress s.GetSmartContractByAddress error")
	}

	// delete transactions from storage
	err = s.transactionStorage.DeleteTransactionsByContractId(sc.ID)
	if err != nil {
		return errors.Wrap(err, "smartcontractstorage: Storage.DeleteSmartContractByAddress s.transactionStorage.DeleteTransactionsByContractId error")
	}

	// delete smartcontract from db
	_, err = s.storage.DB.Exec("DELETE FROM smartcontracts WHERE address = $1", address)
	if err != nil {
		return errors.Wrap(err, "smartcontractstorage: Storage.DeleteSmartContractByAddress s.storage.DB.Exec error")
	}

	return nil
}
