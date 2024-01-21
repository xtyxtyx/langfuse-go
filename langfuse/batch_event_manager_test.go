package langfuse_test

import (
	"context"
	"github.com/wepala/langfuse-go/langfuse"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestBatchEventManager_Enqueue(t *testing.T) {
	t.Run("should enqueue an event", func(t *testing.T) {
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			return NewStringResponse(http.StatusOK, `test`)
		})
		sdk := langfuse.New(nil, langfuse.Options{HttpClient: httpClient})
		eventManager := langfuse.NewBatchEventManager(sdk.Client(), 2, 5)
		if eventManager == nil {
			t.Fatal("expected event manager to be created")
		}
		err := eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}

	})
	t.Run("should fail to enqueue an event if the queue is full", func(t *testing.T) {
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			return NewStringResponse(http.StatusOK, `test`)
		})
		sdk := langfuse.New(nil, langfuse.Options{HttpClient: httpClient, TotalQueues: 1, MaxBatchSize: 1})
		eventManager := sdk.EventManager()
		if eventManager == nil {
			t.Fatal("expected event manager to be created")
		}
		err := eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}
		err = eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err == nil {
			t.Fatalf("expected enqueue to fail")
		}
	})
	t.Run("should enqueue events concurrently", func(t *testing.T) {
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			return NewStringResponse(http.StatusOK, `test`)
		})
		sdk := langfuse.New(nil, langfuse.Options{HttpClient: httpClient})
		eventManager := langfuse.NewBatchEventManager(sdk.Client(), 2, 2)
		if eventManager == nil {
			t.Fatal("expected event manager to be created")
		}
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			err := eventManager.Enqueue("test", "test", map[string]interface{}{})
			if err != nil {
				t.Errorf("expected enqueue to succeed, got %s", err.Error())
				return
			}
		}()
		go func() {
			defer wg.Done()
			err := eventManager.Enqueue("test", "test", map[string]interface{}{})
			if err != nil {
				t.Errorf("expected enqueue to succeed, got %s", err.Error())
			}
		}()
		go func() {
			defer wg.Done()
			err := eventManager.Enqueue("test", "test", map[string]interface{}{})
			if err != nil {
				t.Errorf("expected enqueue to succeed, got %s", err.Error())
			}
		}()
		wg.Wait()
		if eventManager.Queues[1].Events[1] != nil {
			t.Errorf("expected %d event to be enqueued, got an even in the other available slot", 1)
		}
	})
	t.Run("should add to next available queue while one is being processed", func(t *testing.T) {
		t.SkipNow()
		apiCalled := 0
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			time.Sleep(1800 * time.Millisecond)
			apiCalled++
			return NewJsonResponse(http.StatusOK, map[string]interface{}{})
		})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		sdk := langfuse.New(ctx, langfuse.Options{HttpClient: httpClient, TotalQueues: 3, MaxBatchSize: 2})
		eventManager := sdk.EventManager().(*langfuse.BatchEventManager)
		if eventManager == nil {
			t.Fatal("expected event manager to be created")
		}
		if len(eventManager.Queues) != 3 {
			t.Fatalf("expected %d queue to be created, got %d", 3, len(eventManager.Queues))
		}
		err := eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}
		err = eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}
		err = eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}

		success := make(chan bool)
		go eventManager.Process(ctx)
		go func() {
			time.Sleep(100 * time.Millisecond)
			err = eventManager.Enqueue("test", "test", map[string]interface{}{})
			if err != nil {
				t.Errorf("expected enqueue to succeed, got %s", err.Error())
			}
			success <- true
		}()
		select {
		case <-success:
			if len(eventManager.Queues[2].Events) != 1 {
				t.Errorf("expected %d events to be enqueued, got %d", 1, len(eventManager.Queues[2].Events))
			}
		}
	})
}

func TestBatchEventManager_Process(t *testing.T) {
	t.Run("should process events when there are events in the queue", func(t *testing.T) {
		apiCalled := false
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			apiCalled = true
			return NewJsonResponse(http.StatusOK, map[string]interface{}{})
		})
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		sdk := langfuse.New(nil, langfuse.Options{HttpClient: httpClient})
		eventManager := langfuse.NewBatchEventManager(sdk.Client(), 2, 5)
		if eventManager == nil {
			t.Fatal("expected event manager to be created")
		}
		err := eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}
		err = eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}
		go eventManager.Process(ctx)
		time.Sleep(800 * time.Millisecond)
		if !apiCalled {
			t.Errorf("expected api to be called")
		}
	})
	t.Run("should process all queues with events in them", func(t *testing.T) {
		apiCalled := 0
		apiCalls := make(chan int)
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			apiCalled++
			if apiCalled == 2 {
				apiCalls <- apiCalled
			}
			return NewJsonResponse(http.StatusOK, map[string]interface{}{})
		})
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		sdk := langfuse.New(ctx, langfuse.Options{HttpClient: httpClient, TotalQueues: 3, MaxBatchSize: 2})
		eventManager := sdk.EventManager().(*langfuse.BatchEventManager)
		if eventManager == nil {
			t.Fatal("expected event manager to be created")
		}
		err := eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}
		err = eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}
		err = eventManager.Enqueue("test", "test", map[string]interface{}{})
		if err != nil {
			t.Fatalf("expected enqueue to succeed, got %s", err.Error())
		}
		go eventManager.Process(ctx)
		select {
		case <-apiCalls:
			if apiCalled != 2 {
				t.Errorf("expected api to be called %d times, called %d times", 2, apiCalled)
			}
		}
	})
}
