package langfuse

import (
	"context"
	"github.com/wepala/langfuse-go/api"
	"github.com/wepala/langfuse-go/api/client"
	"sync"
	"time"
)

type Queue struct {
	events []interface{}
	mu     sync.Mutex
}

func NewBatchEventManager(client *client.Client, totalQueues int, maxBatchItems int) *BatchEventManager {
	var queues []*Queue
	for i := 0; i < totalQueues; i++ {
		queues = append(queues, &Queue{})
	}
	if maxBatchItems == 0 {
		maxBatchItems = 100
	}
	if totalQueues == 0 {
		totalQueues = 10
	}
	return &BatchEventManager{
		Client:        client,
		queues:        queues,
		maxBatchItems: maxBatchItems,
	}
}

type BatchEventManager struct {
	Client        *client.Client
	queues        []*Queue
	maxBatchItems int
}

func (b *BatchEventManager) Enqueue(id string, eventType string, event interface{}) error {
	//find the next available queue
	var queue *Queue
	//foundQueue := false
	for _, queue = range b.queues {
		queue.mu.Lock()
		if len(queue.events) >= b.maxBatchItems {
			queue.mu.Unlock()
			continue
		}
		queue.events = append(queue.events, map[string]interface{}{
			"id":        id,
			"type":      eventType,
			"body":      event,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		queue.mu.Unlock()
		//foundQueue = true
		break
	}

	return nil
}

func (b *BatchEventManager) Process(ctxt context.Context) {
	var wg sync.WaitGroup

	for {
		select {
		case <-ctxt.Done():
			return
		default:
			var queue *Queue
			for _, queue = range b.queues {
				if len(queue.events) == 0 {
					continue
				}
				wg.Add(1)
				go func(q *Queue) {
					defer wg.Done()
					defer q.mu.Unlock()
					q.mu.Lock()
					resp, err := b.Client.Ingestion.Batch(ctxt, &api.IngestionBatchRequest{Batch: q.events})
					if err != nil {
						//TODO log error
						return
					}
					if len(resp.Errors) > 0 {
						//update the queue to only contain the events that were not sent
						var events []interface{}
						for _, event := range q.events {
							for _, err := range resp.Errors {
								if event.(map[string]interface{})["id"] == err.Id {
									events = append(events, event)
								}
							}
						}
						q.events = events
						return
					}
					q.events = nil

				}(queue)
			}

			// Wait a bit before creating the next goroutine
			time.Sleep(500 * time.Millisecond)
		}
	}

}

func (b *BatchEventManager) Flush(ctxt context.Context) {
	var wg sync.WaitGroup
	var queue *Queue
	for _, queue = range b.queues {
		if len(queue.events) == 0 {
			continue
		}
		wg.Add(1)
		go func(q *Queue) {
			defer wg.Done()
			defer q.mu.Unlock()
			q.mu.Lock()
			resp, err := b.Client.Ingestion.Batch(ctxt, &api.IngestionBatchRequest{Batch: q.events})
			if err != nil {
				//TODO log error
				return
			}
			if len(resp.Errors) > 0 {
				//update the queue to only contain the events that were not sent
				var events []interface{}
				for _, event := range q.events {
					for _, err := range resp.Errors {
						if event.(map[string]interface{})["id"] == err.Id {
							events = append(events, event)
						}
					}
				}
				q.events = events
				return
			}
			q.events = nil

		}(queue)
	}
}
