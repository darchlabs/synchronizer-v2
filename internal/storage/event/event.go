package eventstorage

import (
	"errors"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
)

var (
	ErrEventNotFound = errors.New("event not found error")
)

type Storage struct {
	storage *storage.S
}

func New(s *storage.S) *Storage {
	return &Storage{
		storage: s,
	}
}

func (s *Storage) Stop() error {
	err := s.storage.DB.Close()
	if err != nil {
		return err
	}

	return nil
}

// NOTE:
// 1. InsertEvent Method moved to: internal/storage/event/insert_event_query.go
// 2. UpdateEvent Method moved to: internal/storage/event/update_event_query.go
// 3. ListAllEvents Method moved to: internal/storage/event/select_events_query.go
// 4. ListEvents Method moved to: internal/storage/event/select_events_query.go
// 5. ListEventsByAddress moved to: internal/storage/event/select_events_by_address_query.go
// 6. GetEvent moved to: internal/storage/event/select_event_query.go
// 7. GetEventById Moved to: internal/storage/event/select_event_by_id_query.go
// 8. DeleteEvent Modev to: internal/storage/event/delete_event_query.go
// 9. ListEventData Modev to: internal/storage/event/select_event_data_query.go
// 10. InsertEventData Modev to: internal/storage/event/insert_event_data_query.go
// 11. GetEventsCount Modev to: internal/storage/event/count_events_query.go
// 12. GetEventsCount Modev to: internal/storage/event/count_events_by_address_query.go
// 13. GetEventDataCount Moved to: internal/storage/event/count_event_data_query.go
