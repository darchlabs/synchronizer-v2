package webhook

import (
	"database/sql"
	"encoding/json"
	"time"
)

type WebhookStatus string

const (
	StatusPending   WebhookStatus = "pending"
	StatusFailed    WebhookStatus = "failed"
	StatusDelivered WebhookStatus = "delivered"
)

type WebhookEntityType string

const (
	WebhookEventType       WebhookEntityType = "event"
	WebhookTransactionType WebhookEntityType = "transaction"
)

type Webhook struct {
	ID          string            `db:"id" json:"id"`
	EntityType  WebhookEntityType `db:"entity_type" json:"entity_type"`
	EntityID    string            `db:"entity_id" json:"entity_id"`
	Endpoint    string            `db:"endpoint" json:"endpoint" validate:"url"`
	Payload     json.RawMessage   `db:"payload" json:"payload"`
	MaxAttempts int               `db:"max_attempts" json:"max_attempts"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
	SentAt      sql.NullTime      `db:"sent_at" json:"sent_at"`
	Attempts    int               `db:"attempts" json:"attempts"`
	NextRetryAt sql.NullTime      `db:"next_retry_at" json:"next_retry_at"`
	Status      WebhookStatus     `db:"status" json:"status"`
}

func (w *Webhook) ToWebhookEventResponse() *WebhookResponse {
	return &WebhookResponse{
		ID:        w.ID,
		Type:      string(w.EntityType),
		Endpoint:  w.Endpoint,
		Payload:   w.Payload,
		CreatedAt: w.CreatedAt,
	}
}

type WebhookEventPayload struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	BlockNumber int64           `json:"block_number"`
	Tx          string          `json:"tx"`
	Data        json.RawMessage `json:"data"`
}

type WebhookResponse struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Endpoint  string          `json:"endpoint"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}
