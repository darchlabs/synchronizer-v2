package webhooksender

import (
	"github.com/darchlabs/synchronizer-v2/pkg/webhook"
)

type WebhookPriorityQueue struct {
	webhooks []*webhook.Webhook
}

func NewWebhookPriorityQueue() *WebhookPriorityQueue {
	return &WebhookPriorityQueue{
		webhooks: make([]*webhook.Webhook, 0),
	}
}

func (pq *WebhookPriorityQueue) Len() int {
	return len(pq.webhooks)
}

func (pq *WebhookPriorityQueue) Less(i, j int) bool {
	return pq.webhooks[i].Status != webhook.StatusFailed && pq.webhooks[j].Status == webhook.StatusFailed
}

func (pq *WebhookPriorityQueue) Swap(i, j int) {
	pq.webhooks[i], pq.webhooks[j] = pq.webhooks[j], pq.webhooks[i]
}

func (pq *WebhookPriorityQueue) Push(w *webhook.Webhook) {
	pq.webhooks = append(pq.webhooks, w)
}

func (pq *WebhookPriorityQueue) Pop() *webhook.Webhook {
	old := pq.webhooks
	n := len(old)
	x := old[n-1]
	pq.webhooks = old[0 : n-1]
	return x
}
