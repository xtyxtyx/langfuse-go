package langfuse_test

import (
	"context"
	"github.com/wepala/langfuse-go/langfuse"
	"testing"
)

func TestTrace_Span(t *testing.T) {
	t.Run("should create a span with the trace id set", func(t *testing.T) {
		enqueueCalls := 0
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				enqueueCalls++
				if enqueueCalls == 2 {
					var ok bool
					var span *langfuse.Span
					if eventType != "span-create" {
						t.Errorf("expected event type to be %s, got %s", "span-create", eventType)
					}

					if span, ok = tevent.(*langfuse.Span); !ok || span == nil {
						t.Errorf("expected event to be a trace")
					}
					//check that an id is set
					if span.ID == "" {
						t.Errorf("expected event id to be set")
					}
					if span.ParentID != "test" {
						t.Errorf("expected parentObservationId to be set to %s, got %s", "test", span.ParentID)
					}

					if span.TraceID != "test" {
						t.Errorf("expected traceId to be set to %s, got %s", "test", span.TraceID)
					}
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		trace, _ := sdk.Trace(context.Background(), &langfuse.Trace{
			BasicObservation: langfuse.BasicObservation{ID: "test"},
		})
		if trace == nil {
			t.Fatalf("expected trace to be created")
		}
		trace.Span(&langfuse.Span{})
		if len(eventManager.calls.Enqueue) != 2 {
			t.Errorf("expected  %d events to be enqueued, got %d", 2, len(eventManager.calls.Enqueue))
		}
	})
}
