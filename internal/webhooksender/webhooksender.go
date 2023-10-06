package webhooksender

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	webhookstorage "github.com/darchlabs/synchronizer-v2/internal/storage/webhook"
	"github.com/darchlabs/synchronizer-v2/pkg/webhook"
	"github.com/pkg/errors"
)

type WebhookSender struct {
	WebhookStorage *webhookstorage.Storage
	HTTPClient     *http.Client
	TickerTime     time.Duration
	webhookQueue   *WebhookPriorityQueue
	InQueue        map[string]struct{}
}

func NewWebhookSender(storage *webhookstorage.Storage, client *http.Client, tickerTime time.Duration) *WebhookSender {
	return &WebhookSender{
		WebhookStorage: storage,
		HTTPClient:     client,
		TickerTime:     tickerTime,
		webhookQueue:   NewWebhookPriorityQueue(),
		InQueue:        make(map[string]struct{}),
	}
}

func (s *WebhookSender) EnqueueWebhook(wh *webhook.Webhook) {
	s.webhookQueue.Push(wh)
	s.InQueue[wh.ID] = struct{}{}
}

func (s *WebhookSender) ProcessWebhooks() {
	for {
		if s.webhookQueue.Len() == 0 {
			time.Sleep(s.TickerTime * time.Second)
			continue
		}

		wh := s.webhookQueue.Pop()

		err := s.SendWebhook(wh)
		if err != nil {
			wh.Status = webhook.StatusFailed
			wh.NextRetryAt = sql.NullTime{Time: time.Now().Add(s.TickerTime * time.Second), Valid: true}
		} else {
			wh.Status = webhook.StatusDelivered
			wh.SentAt = sql.NullTime{Time: time.Now(), Valid: true}
		}
		wh.UpdatedAt = time.Now()
		wh.Attempts++

		if _, err = s.WebhookStorage.UpdateWebhook(wh); err != nil {
			log.Fatalf("Fatal error updating webhook in the database: %s\n", err)
		}

		delete(s.InQueue, wh.ID)
	}
}

func (s *WebhookSender) SendWebhook(wh *webhook.Webhook) error {
	// parse webhook to bytes
	b, err := json.Marshal(wh.ToWebhookEventResponse())
	if err != nil {
		return err
	}

	// send POST using the webhook
	fmt.Println("~~~~~ Before Send Webhook")
	fmt.Println(" body: ", string(b))
	req, err := http.NewRequest("POST", wh.Endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	_, err = s.HTTPClient.Do(req)
	fmt.Println("~~~~~ After Webhook is sent")

	return err
}

func (s *WebhookSender) StartRetries() {
	for {
		webhooks, err := s.WebhookStorage.GetWebhooksForRetry(s.InQueue)
		if err != nil {
			time.Sleep(s.TickerTime * time.Second)
			continue
		}

		if len(webhooks) == 0 {
			time.Sleep(s.TickerTime * time.Second)
			continue
		}

		for _, wh := range webhooks {
			err := s.SendWebhook(wh)
			if err != nil {
				wh.Status = webhook.StatusFailed
				wh.NextRetryAt = sql.NullTime{Time: time.Now().Add(s.TickerTime * time.Second), Valid: true}
			} else {
				wh.Status = webhook.StatusDelivered
				wh.SentAt = sql.NullTime{Time: time.Now(), Valid: true}
			}
			wh.UpdatedAt = time.Now()
			wh.Attempts++

			if _, err = s.WebhookStorage.UpdateWebhook(wh); err != nil {
				log.Fatalf("Fatal error updating webhook in the database: %s\n", err)
			}
		}
	}
}

func (s *WebhookSender) CreateAndSendWebhook(wh *webhook.Webhook) error {
	// Create the webhook in the database
	wh, err := s.WebhookStorage.CreateWebhook(wh)
	if errors.Is(err, webhookstorage.DuplicatedWebhookErr) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "webhooksender: error creating webhook")
	}

	// Enqueue the webhook for sending
	s.EnqueueWebhook(wh)

	return nil
}

func (s *WebhookSender) InitializeFromStorage() error {
	webhooks, err := s.WebhookStorage.GetQueuedWebhooks()
	if err != nil {
		return errors.Wrap(err, "webhooksender: error retrieving queued webhooks from storage")
	}

	for _, wh := range webhooks {
		s.EnqueueWebhook(wh)
	}

	return nil
}
