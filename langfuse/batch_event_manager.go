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
	id        int
	Events    []interface{}
	nextEntry int
	mu        sync.Mutex
	maxItems  int
}

func (q *Queue) Reset() {
	q.Events = make([]interface{}, q.maxItems)
	q.nextEntry = 0
}

func NewBatchEventManager(client *client.Client, totalQueues int, maxBatchItems int) *BatchEventManager {
	var queues []*Queue

	if maxBatchItems == 0 {
		maxBatchItems = 100
	}
	if totalQueues == 0 {
		totalQueues = 10
	}

	for i := 0; i < totalQueues; i++ {
		queues = append(queues, &Queue{id: i, Events: make([]interface{}, maxBatchItems), maxItems: maxBatchItems})
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
		if queue.nextEntry == b.maxBatchItems {
			continue
		}
		queue.Events[queue.nextEntry] = map[string]interface{}{
			"id":        id,
			"type":      eventType,
			"body":      event,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		queue.nextEntry++
		log.Printf("add to queue %d,queue length %d, max %d", queue.id, queue.nextEntry, b.maxBatchItems)
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
		if queue.nextEntry == 0 {
			continue
		}
		wg.Add(1)
		go func(q *Queue) {
			defer wg.Done()
			defer q.mu.Unlock()
			q.mu.Lock()
			if q.nextEntry == 0 {
				return
			}
			resp, err := b.Client.Ingestion.Batch(ctxt, &api.IngestionBatchRequest{Batch: q.Events[:q.nextEntry]})
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
			q.Reset()
		}(queue)
	}
}
