package langfuse

import (
	"context"
	"fmt"
	"github.com/wepala/langfuse-go/api"
	"github.com/wepala/langfuse-go/api/client"
	"log"
	"sync"
	"time"
)

type Queue struct {
	id     int
	Events []interface{}
	mu     sync.Mutex
}

func NewBatchEventManager(client *client.Client, totalQueues int, maxBatchItems int) *BatchEventManager {
	var queues []*Queue
	for i := 0; i < totalQueues; i++ {
		queues = append(queues, &Queue{id: i})
	}
	if maxBatchItems == 0 {
		maxBatchItems = 100
	}
	if totalQueues == 0 {
		totalQueues = 10
	}
	return &BatchEventManager{
		Client:        client,
		Queues:        queues,
		maxBatchItems: maxBatchItems,
	}
}

type BatchEventManager struct {
	Client        *client.Client
	Queues        []*Queue
	maxBatchItems int
}

func (b *BatchEventManager) Enqueue(id string, eventType string, event interface{}) error {
	//find the next available queue
	var queue *Queue
	for _, queue = range b.Queues {
		queue.mu.Lock()
		if len(queue.Events) >= b.maxBatchItems {
			continue
		}
		queue.Events = append(queue.Events, map[string]interface{}{
			"id":        id,
			"type":      eventType,
			"body":      event,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		log.Printf("add to queue %d,queue length %d, max %d", queue.id, len(queue.Events), b.maxBatchItems)
		queue.mu.Unlock()
		return nil
	}

	return fmt.Errorf("no queue available")
}

func (b *BatchEventManager) Process(ctxt context.Context) {
	for {
		select {
		case <-ctxt.Done():
			return
		default:
			b.Flush(ctxt)
			// Wait a bit before creating the next goroutine
			time.Sleep(500 * time.Millisecond)
		}
	}

}

func (b *BatchEventManager) Flush(ctxt context.Context) {
	var wg sync.WaitGroup
	var queue *Queue
	for _, queue = range b.Queues {
		if len(queue.Events) == 0 {
			continue
		}
		wg.Add(1)
		go func(q *Queue) {
			defer wg.Done()
			defer q.mu.Unlock()
			q.mu.Lock()
			resp, err := b.Client.Ingestion.Batch(ctxt, &api.IngestionBatchRequest{Batch: q.Events})
			if err != nil {
				//TODO log error
				return
			}
			if len(resp.Errors) > 0 {
				//update the queue to only contain the events that were not sent
				var events []interface{}
				for _, event := range q.Events {
					for _, err := range resp.Errors {
						if event.(map[string]interface{})["id"] == err.Id {
							events = append(events, event)
						}
					}
				}
				q.Events = events
				return
			}
			q.Events = nil
		}(queue)
	}
}
