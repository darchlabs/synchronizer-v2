package webhookstorage

import (
	"database/sql"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/webhook"
	"github.com/pkg/errors"
)

var DuplicatedWebhookErr = errors.New("webhookstorage: duplicated webhook")

type Storage struct {
	storage *storage.S
}

func New(s *storage.S) *Storage {
	return &Storage{
		storage: s,
	}
}

func (s *Storage) CreateWebhook(wh *webhook.Webhook) (*webhook.Webhook, error) {
	selectWebhookQuery := `SELECT * FROM webhooks WHERE user_id = $1 AND tx = $2`
	whs := make([]*webhook.Webhook, 0)
	s.storage.DB.Select(&whs, selectWebhookQuery, wh.UserID, wh.Tx)
	// by the moment we can omit the error because is used as check for dup webhooks
	if len(whs) > 0 {
		return nil, DuplicatedWebhookErr
	}

	inserWebhookQuery := `
		INSERT INTO webhooks (id, user_id, tx, entity_type, entity_id, endpoint, payload, created_at, updated_at, sent_at, next_retry_at) 
		VALUES (:id, :user_id, :tx, :entity_type, :entity_id, :endpoint, :payload, :created_at, :updated_at, :sent_at, :next_retry_at)
		RETURNING id
	`

	rows, err := s.storage.DB.NamedQuery(inserWebhookQuery, wh)
	if err != nil {
		return nil, errors.Wrap(err, "webhookstorage: error creating webhook")
	}
	defer rows.Close()

	if rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "webhookstorage: error scanning webhook ID")
		}
		wh.ID = id
	} else {
		return nil, errors.New("webhookstorage: no ID returned after webhook creation")
	}

	createdWebhook, err := s.GetWebhookByID(wh.ID)
	if err != nil {
		return nil, err
	}

	return createdWebhook, nil
}

func (s *Storage) UpdateWebhook(wh *webhook.Webhook) (*webhook.Webhook, error) {
	query := `
		UPDATE webhooks 
		SET endpoint = :endpoint, payload = :payload, status = :status, attempts = :attempts, next_retry_at = :next_retry_at, updated_at = :updated_at, sent_at = :sent_at 
		WHERE id = :id
	`

	_, err := s.storage.DB.NamedExec(query, wh)
	if err != nil {
		return nil, errors.Wrap(err, "webhookstorage: error updating webhook")
	}

	updatedWebhook, err := s.GetWebhookByID(wh.ID)
	if err != nil {
		return nil, err
	}

	return updatedWebhook, nil
}

func (s *Storage) ListAllWebhooks() ([]*webhook.Webhook, error) {
	// define webhooks response
	webhooks := []*webhook.Webhook{}

	// get webhooks from db
	whQuery := "SELECT * FROM webhooks"
	err := s.storage.DB.Select(&webhooks, whQuery)
	if err != nil {
		return nil, err
	}

	return webhooks, nil
}

func (s *Storage) GetWebhookByID(id string) (*webhook.Webhook, error) {
	// get webhook from db
	wh := &webhook.Webhook{}
	err := s.storage.DB.Get(wh, "SELECT * FROM webhooks WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return wh, nil
}

func (s *Storage) ListWebhooks(smartcontractID string) ([]*webhook.Webhook, error) {
	webhooks := []*webhook.Webhook{}
	err := s.storage.DB.Select(&webhooks, "SELECT * FROM webhooks WHERE smartcontract_id = $1", smartcontractID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*webhook.Webhook{}, nil
		}
		return nil, err
	}

	return webhooks, nil
}

func (s *Storage) GetWebhooksForRetry(inQueue map[string]struct{}) ([]*webhook.Webhook, error) {
	webhooks := []*webhook.Webhook{}
	rows, err := s.storage.DB.Query(
		"SELECT * FROM webhooks WHERE (status = $1 OR status = $2) AND next_retry_at <= $3 AND attempts < max_attempts;",
		webhook.StatusFailed, webhook.StatusPending, time.Now())
	if err != nil {
		if err == sql.ErrNoRows {
			return []*webhook.Webhook{}, nil
		}

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		wh := &webhook.Webhook{}
		if err := rows.Scan(
			&wh.ID,
			&wh.EntityType,
			&wh.EntityID,
			&wh.Endpoint,
			&wh.Payload,
			&wh.MaxAttempts,
			&wh.CreatedAt,
			&wh.UpdatedAt,
			&wh.SentAt,
			&wh.Attempts,
			&wh.Status,
			&wh.NextRetryAt,
		); err != nil {
			return nil, err
		}

		// check if webhook are in queue
		if _, ok := inQueue[wh.ID]; !ok {
			webhooks = append(webhooks, wh)
		}
	}

	return webhooks, nil
}

func (s *Storage) GetQueuedWebhooks() ([]*webhook.Webhook, error) {
	webhooks := []*webhook.Webhook{}
	query := "SELECT * FROM webhooks WHERE status = $1"

	err := s.storage.DB.Select(&webhooks, query, webhook.StatusPending)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*webhook.Webhook{}, nil
		}
		return nil, err
	}

	return webhooks, nil
}
