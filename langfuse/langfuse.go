package langfuse

import (
	"context"
	"github.com/segmentio/ksuid"
	"github.com/wepala/langfuse-go/api/client"
	"net/http"
	"os"
	"time"
)

type Options struct {
	HttpClient   *http.Client
	EventManager EventManager `json:"-"`
	PublicKey    string       `json:"-"`
	SecretKey    string       `json:"-"`
	Host         string       `json:"host"`
	Release      string       `json:"release"`
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

func (l *LangFuse) Trace(ctxt context.Context, opts *Trace) *Trace {
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

	l.eventManager.Enqueue(opts.ID, TRACE_CREATE, opts)
	return opts
}

func (l *LangFuse) Span(ctxt context.Context, opts *Span) *Span {
	if opts == nil {
		opts = &Span{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}

	l.eventManager.Enqueue(opts.ID, SPAN_CREATE, opts)
	return opts
}

func (l *LangFuse) Event(ctxt context.Context, opts *Event) *Event {
	if opts == nil {
		opts = &Event{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}

	l.eventManager.Enqueue(opts.ID, EVENT_CREATE, opts)
	return opts
}

func (l *LangFuse) Generation(ctxt context.Context, opts *Generation) *Generation {
	if opts == nil {
		opts = &Generation{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}

	l.eventManager.Enqueue(opts.ID, GENERATION_CREATE, opts)
	return opts
}

func (l *LangFuse) Score(ctxt context.Context, opts *Score) *Score {
	if opts == nil {
		opts = &Score{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.TraceID == "" {
		return nil
	}

	if opts.Name == "" {
		return nil
	}

	l.eventManager.Enqueue(opts.ID, SCORE_CREATE, opts)
	return opts
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
		batchEventManager = NewBatchEventManager(tclient, 10, 0)
		options.EventManager = batchEventManager
	}

	lf := &LangFuse{
		client:       tclient,
		eventManager: options.EventManager,
	}

	if batchEventManager != nil {
		if ctxt == nil {
			ctxt = context.Background()
		}
		tctxt, cancel := context.WithCancel(ctxt)
		go batchEventManager.Process(tctxt)
		lf.Shutdown = cancel
	}

	return lf
}
