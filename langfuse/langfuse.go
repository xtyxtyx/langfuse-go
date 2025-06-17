package langfuse

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/wepala/langfuse-go/api/client"
)

type Options struct {
	HttpClient   *http.Client
	EventManager EventManager `json:"-"`
	PublicKey    string       `json:"-"`
	SecretKey    string       `json:"-"`
	Host         string       `json:"host"`
	Release      string       `json:"release"`
	TotalQueues  int          `json:"total_queues"`
	MaxBatchSize int          `json:"max_batch_size"`
}

type LangFuse struct {
	client       *client.Client
	eventManager EventManager
	Shutdown     context.CancelFunc
}

func (l *LangFuse) Client() *client.Client {
	return l.client
}

func (l *LangFuse) EventManager() EventManager {
	return l.eventManager
}

func (l *LangFuse) Trace(ctxt context.Context, opts *Trace) (*Trace, error) {
	if opts == nil {
		opts = &Trace{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.Release == "" {
		opts.Release = os.Getenv("LANGFUSE_RELEASE")
	}

	opts.eventManager = l.eventManager

	err := l.eventManager.Enqueue("", TRACE_CREATE, opts)
	return opts, err
}

func (l *LangFuse) Span(ctxt context.Context, opts *Span) (*Span, error) {
	if opts == nil {
		opts = &Span{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}

	l.eventManager.Enqueue("", SPAN_CREATE, opts)
	return opts, nil
}

func (l *LangFuse) Event(ctxt context.Context, opts *Event) (*Event, error) {
	if opts == nil {
		opts = &Event{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}

	l.eventManager.Enqueue("", EVENT_CREATE, opts)
	return opts, nil
}

func (l *LangFuse) Generation(ctxt context.Context, opts *Generation) (*Generation, error) {
	if opts == nil {
		opts = &Generation{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}

	l.eventManager.Enqueue("", GENERATION_CREATE, opts)
	return opts, nil
}

func (l *LangFuse) Score(ctxt context.Context, opts *Score) (*Score, error) {
	if opts == nil {
		opts = &Score{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.TraceID == "" {
		return nil, fmt.Errorf("trace id is required")
	}

	if opts.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	l.eventManager.Enqueue("", SCORE_CREATE, opts)
	return opts, nil
}

func (l *LangFuse) Start(ctxt context.Context) {
	if l.eventManager != nil {
		if ctxt == nil {
			ctxt = context.Background()
		}
		tctxt, cancel := context.WithCancel(ctxt)
		if batchEventManager, ok := l.eventManager.(*BatchEventManager); ok {
			go batchEventManager.Process(tctxt)
		}
		l.Shutdown = cancel
	}
}

func New(ctxt context.Context, options Options) *LangFuse {
	if options.PublicKey == "" {
		options.PublicKey = os.Getenv("LANGFUSE_PUBLIC_KEY")
	}
	if options.SecretKey == "" {
		options.SecretKey = os.Getenv("LANGFUSE_SECRET_KEY")
	}

	if options.Host == "" {
		options.Host = os.Getenv("LANGFUSE_HOST")
		if options.Host == "" {
			options.Host = "https://cloud.langfuse.com"
		}
	}

	tclient := client.NewClient(client.WithBaseURL(options.Host), client.WithHTTPClient(options.HttpClient), client.WithAuthBasic(options.PublicKey, options.SecretKey))

	var batchEventManager *BatchEventManager
	if options.EventManager == nil {
		if options.TotalQueues == 0 {
			options.TotalQueues = 10
		}
		if options.MaxBatchSize == 0 {
			options.MaxBatchSize = 100
		}
		batchEventManager = NewBatchEventManager(tclient, options.TotalQueues, options.MaxBatchSize)
		options.EventManager = batchEventManager
	}

	lf := &LangFuse{
		client:       tclient,
		eventManager: options.EventManager,
	}
	return lf
}
