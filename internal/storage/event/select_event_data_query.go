package eventstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/pkg/errors"
)

func (s *Storage) ListEventData(address string, eventName string, sort string, limit int64, offset int64) ([]*event.EventData, error) {
	// define events data response
	eventsData := []*event.EventData{}

	// define and make the query on db
	err := s.storage.DB.Select(
		&eventsData,
		fmt.Sprintf(`
			SELECT event_data.*
			FROM event_data
			JOIN event
			ON event_data.event_id = event.id
			JOIN abi ON event.abi_id = abi.id
			WHERE event.address = $1
			AND abi.name = $2
			ORDER BY event_data.created_at %s
			LIMIT $3
			OFFSET $4;`,
			sort),
		address,
		eventName,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.ListEventData s.storage.DB.Select error")
	}

	return eventsData, nil
}
