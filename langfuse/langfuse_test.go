package langfuse_test

import (
	"context"
	"encoding/base64"
	"github.com/wepala/langfuse-go/langfuse"
	"net/http"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("should use environment variables to setup api client if no options are provided", func(t *testing.T) {
		_ = os.Setenv("LANGFUSE_PUBLIC_KEY", "public-key")
		_ = os.Setenv("LANGFUSE_SECRET_KEY", "secret-key")
		_ = os.Setenv("LANGFUSE_HOST", "http://localhost:8080")
		apiCalled := false

		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			apiCalled = true
			//check that the expected headers are sent
			if req.Header.Get("Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte("public-key: secret-key")) {
				t.Errorf("expected Authorization header to be set")
			}
			if req.URL.Host != "localhost:8080" {
				t.Errorf("expected host to be set to %s, got %s", "localhost:8080", req.URL.Host)
			}
			return NewStringResponse(http.StatusOK, `test`)
		})
		sdk := langfuse.New(nil, langfuse.Options{HttpClient: httpClient})
		if sdk == nil {
			t.Fatal("expected sdk to be created")
		}
		client := sdk.Client()
		if client == nil {
			t.Fatal("expected client to be created")
		}
		//call the health endpoint and confirm that the expected headers are sent
		_, _ = client.Health.Health(context.TODO())
		if !apiCalled {
			t.Errorf("expected api to be called")
		}
	})
	t.Run("should fall back to use cloud.langfuse.com if no host is provided in the options of environment", func(t *testing.T) {
		_ = os.Setenv("LANGFUSE_HOST", "")
		apiCalled := false
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			apiCalled = true
			if req.URL.Host != "cloud.langfuse.com" {
				t.Errorf("expected host to be set to %s, got %s", "cloud.langfuse.com", req.URL.Host)
			}
			return NewStringResponse(http.StatusOK, `test`)
		})
		sdk := langfuse.New(nil, langfuse.Options{HttpClient: httpClient})
		if sdk == nil {
			t.Fatal("expected sdk to be created")
		}
		client := sdk.Client()
		if client == nil {
			t.Fatal("expected client to be created")
		}
		//call the health endpoint and confirm that the expected headers are sent
		_, _ = client.Health.Health(context.TODO())
		if !apiCalled {
			t.Errorf("expected api to be called")
		}
	})
	t.Run("should create an event manager with the api client", func(t *testing.T) {

		sdk := langfuse.New(nil, langfuse.Options{})
		if sdk == nil {
			t.Fatal("expected sdk to be created")
		}
		client := sdk.Client()
		if client == nil {
			t.Fatal("expected client to be created")
		}
		eventManager := sdk.EventManager()
		if eventManager == nil {
			t.Fatal("expected event manager to be created")
		}
	})

}

func TestLangFuse_Trace(t *testing.T) {
	t.Run("should return a trace object with a default id and default release and add trace event to queue", func(t *testing.T) {
		os.Setenv("LANGFUSE_RELEASE", "default release")
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				var ok bool
				var trace *langfuse.Trace

				if eventType != langfuse.TRACE_CREATE {
					t.Errorf("expected event type to be %s, got %s", langfuse.TRACE_CREATE, eventType)
				}

				if trace, ok = tevent.(*langfuse.Trace); !ok || trace == nil {
					t.Errorf("expected event to be a trace")
				}
				//check that an id is set
				if trace.ID == "" {
					t.Errorf("expected event id to be set")
				}

				if trace.Release != "default release" {
					t.Errorf("expected event release to be set to default release, got %s", trace.Release)
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		trace := sdk.Trace(context.Background(), &langfuse.Trace{})
		if trace == nil {
			t.Errorf("expected trace to be created")
		}
		if len(eventManager.calls.Enqueue) != 1 {
			t.Errorf("expected event to be added to queue")
		}
	})
	t.Run("should return a trace object with the id that was set", func(t *testing.T) {
		os.Setenv("LANGFUSE_RELEASE", "default release")
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				var ok bool
				var trace *langfuse.Trace

				if trace, ok = tevent.(*langfuse.Trace); !ok || trace == nil {
					t.Errorf("expected event to be a trace")
				}

				//check that an id is set
				if trace.ID != "test" {
					t.Errorf("expected event id to be set to %s, got %s", "test", trace.ID)
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		trace := sdk.Trace(context.Background(), &langfuse.Trace{
			BasicObservation: langfuse.BasicObservation{
				ID: "test",
			},
		})
		if trace == nil {
			t.Errorf("expected trace to be created")
		}
		if len(eventManager.calls.Enqueue) != 1 {
			t.Errorf("expected event to be added to queue")
		}
	})
}

func TestLangFuse_Span(t *testing.T) {
	t.Run("should return a span object with a default id and add span event to queue", func(t *testing.T) {
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				var ok bool
				var span *langfuse.Span

				if eventType != langfuse.SPAN_CREATE {
					t.Errorf("expected event type to be %s, got %s", langfuse.SPAN_CREATE, eventType)
				}

				if span, ok = tevent.(*langfuse.Span); !ok || span == nil {
					t.Errorf("expected event to be a span")
				}
				//check that an id is set
				if span.ID == "" {
					t.Errorf("expected event id to be set")
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		span := sdk.Span(context.Background(), &langfuse.Span{})
		if span == nil {
			t.Errorf("expected span to be created")
		}
		if len(eventManager.calls.Enqueue) != 1 {
			t.Errorf("expected event to be added to queue")
		}
	})

	t.Run("should return a span object with the id that was set", func(t *testing.T) {
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				var ok bool
				var span *langfuse.Span

				if span, ok = tevent.(*langfuse.Span); !ok || span == nil {
					t.Errorf("expected event to be a span")
				}

				//check that an id is set
				if span.ID != "test" {
					t.Errorf("expected event id to be set to %s, got %s", "test", span.ID)
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		span := sdk.Span(context.Background(), &langfuse.Span{
			BasicObservation: langfuse.BasicObservation{
				ID: "test",
			},
		})
		if span == nil {
			t.Errorf("expected span to be created")
		}
		if len(eventManager.calls.Enqueue) != 1 {
			t.Errorf("expected event to be added to queue")
		}
	})
}

func TestLangFuse_Generation(t *testing.T) {
	t.Run("should return a generation object with a default id and add generation event to queue", func(t *testing.T) {
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				var ok bool
				var generation *langfuse.Generation

				if eventType != langfuse.GENERATION_CREATE {
					t.Errorf("expected event type to be %s, got %s", langfuse.GENERATION_CREATE, eventType)
				}

				if generation, ok = tevent.(*langfuse.Generation); !ok || generation == nil {
					t.Errorf("expected event to be a generation")
				}
				//check that an id is set
				if generation.ID == "" {
					t.Errorf("expected event id to be set")
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		generation := sdk.Generation(context.Background(), &langfuse.Generation{})
		if generation == nil {
			t.Errorf("expected generation to be created")
		}
		if len(eventManager.calls.Enqueue) != 1 {
			t.Errorf("expected event to be added to queue")
		}
	})

	t.Run("should return a generation object with the id that was set", func(t *testing.T) {
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				var ok bool
				var generation *langfuse.Generation

				if generation, ok = tevent.(*langfuse.Generation); !ok || generation == nil {
					t.Errorf("expected event to be a generation")
				}

				//check that an id is set
				if generation.ID != "test" {
					t.Errorf("expected event id to be set to %s, got %s", "test", generation.ID)
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		generation := sdk.Generation(context.Background(), &langfuse.Generation{
			BasicObservation: langfuse.BasicObservation{
				ID: "test",
			},
		})
		if generation == nil {
			t.Errorf("expected generation to be created")
		}
		if len(eventManager.calls.Enqueue) != 1 {
			t.Errorf("expected event to be added to queue")
		}
	})
}

func TestLangFuse_Score(t *testing.T) {
	t.Run("should return a score object with a default id and add score event to queue", func(t *testing.T) {
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				var ok bool
				var score *langfuse.Score

				if eventType != langfuse.SCORE_CREATE {
					t.Errorf("expected event type to be %s, got %s", langfuse.SCORE_CREATE, eventType)
				}

				if score, ok = tevent.(*langfuse.Score); !ok || score == nil {
					t.Errorf("expected event to be a score")
				}
				//check that an id is set
				if score.ID == "" {
					t.Errorf("expected event id to be set")
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		score := sdk.Score(context.Background(), &langfuse.Score{
			BasicObservation: langfuse.BasicObservation{
				Name:    "test",
				TraceID: "test",
			},
		})
		if score == nil {
			t.Errorf("expected score to be created")
		}
		if len(eventManager.calls.Enqueue) != 1 {
			t.Errorf("expected event to be added to queue")
		}
	})

	t.Run("should return a score object with the id that was set", func(t *testing.T) {
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				var ok bool
				var score *langfuse.Score

				if score, ok = tevent.(*langfuse.Score); !ok || score == nil {
					t.Errorf("expected event to be a score")
				}

				//check that an id is set
				if score.ID != "test" {
					t.Errorf("expected event id to be set to %s, got %s", "test", score.ID)
				}

				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		score := sdk.Score(context.Background(), &langfuse.Score{
			BasicObservation: langfuse.BasicObservation{
				ID:      "test",
				Name:    "test",
				TraceID: "test",
			},
		})
		if score == nil {
			t.Errorf("expected score to be created")
		}
		if len(eventManager.calls.Enqueue) != 1 {
			t.Errorf("expected event to be added to queue")
		}
	})

	t.Run("should return nil if trace id is not set", func(t *testing.T) {
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				t.Errorf("expected event not to be added to queue")
				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		score := sdk.Score(context.Background(), &langfuse.Score{
			BasicObservation: langfuse.BasicObservation{
				ID:   "test",
				Name: "test",
			},
		})
		if score != nil {
			t.Errorf("expected score not to be created")
		}
		if len(eventManager.calls.Enqueue) != 0 {
			t.Errorf("expected event not to be added to queue")
		}
	})

	t.Run("should return nil if name is not set", func(t *testing.T) {
		eventManager := &EventManagerMock{
			EnqueueFunc: func(id string, eventType string, tevent interface{}) error {
				t.Errorf("expected event not to be added to queue")
				return nil
			},
		}
		sdk := langfuse.New(nil, langfuse.Options{
			EventManager: eventManager,
		})
		score := sdk.Score(context.Background(), &langfuse.Score{
			BasicObservation: langfuse.BasicObservation{
				ID:      "test",
				TraceID: "test",
			},
		})
		if score != nil {
			t.Errorf("expected score not to be created")
		}
		if len(eventManager.calls.Enqueue) != 0 {
			t.Errorf("expected event not to be added to queue")
		}
	})
}
