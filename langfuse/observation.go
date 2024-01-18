package langfuse

import (
	"errors"
	"github.com/segmentio/ksuid"
	"time"
)

type Observation interface {
	Span(opts *Span) *Span
	Event(opts *Event) *Event
	Generation(opts *Generation) *Generation
	Score(opts *Score) *Score
}

type BasicObservation struct {
	ID            string                 `json:"id,omitempty"`
	Name          string                 `json:"name,omitempty"`
	TraceID       string                 `json:"traceId,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Level         string                 `json:"level,omitempty"`
	StatusMessage string                 `json:"statusMessage,omitempty"`
	Input         map[string]interface{} `json:"input,omitempty"`
	Output        map[string]interface{} `json:"output,omitempty"`
	Version       string                 `json:"version,omitempty"`
	ParentID      string                 `json:"parentObservationId,omitempty"`
	eventManager  EventManager
}

func (o BasicObservation) Span(span *Span) *Span {
	if span == nil {
		span = &Span{}
	}

	if span.ID == "" {
		span.ID = ksuid.New().String()
	}

	if span.ParentID == "" {
		span.ParentID = o.ID
	}

	if span.TraceID == "" {
		span.TraceID = o.TraceID
	}

	if span.StartTime.IsZero() {
		span.StartTime = time.Now()
	}

	span.eventManager = o.eventManager
	o.eventManager.Enqueue(span.ID, SPAN_CREATE, span)

	return span
}

func (o BasicObservation) Event(opts *Event) *Event {
	if opts == nil {
		opts = &Event{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.ParentID == "" {
		opts.ParentID = o.ID
	}

	if opts.TraceID == "" {
		opts.TraceID = o.TraceID
	}

	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}

	opts.eventManager = o.eventManager
	o.eventManager.Enqueue(opts.ID, EVENT_CREATE, opts)

	return opts
}

func (o BasicObservation) Generation(generation *Generation) *Generation {
	if generation == nil {
		generation = &Generation{}
	}

	if generation.ID == "" {
		generation.ID = ksuid.New().String()
	}

	if generation.ParentID == "" {
		generation.ParentID = o.ID
	}

	if generation.TraceID == "" {
		generation.TraceID = o.TraceID
	}

	if generation.StartTime.IsZero() {
		generation.StartTime = time.Now()
	}

	generation.eventManager = o.eventManager
	o.eventManager.Enqueue(generation.ID, GENERATION_CREATE, generation)

	return generation
}

func (o BasicObservation) Score(opts *Score) *Score {
	if opts == nil {
		opts = &Score{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	if opts.TraceID == "" {
		opts.TraceID = o.TraceID
	}

	if opts.Name == "" {
		opts.Name = o.Name
	}

	opts.eventManager = o.eventManager
	o.eventManager.Enqueue(opts.ID, SCORE_CREATE, opts)

	return opts
}

type Span struct {
	BasicObservation
	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime,omitempty"`
}

func (s *Span) End() error {
	if s.ID == "" {
		return errors.New("span id is not set")
	}
	s.EndTime = time.Now()
	s.eventManager.Enqueue(s.ID, SPAN_UPDATE, s)
	return nil
}

type Generation struct {
	BasicObservation
	CompletionStartTime time.Time              `json:"completionStartTime,omitempty"`
	Model               string                 `json:"model,omitempty"`
	Usage               map[string]interface{} `json:"usage,omitempty"`
	ModelParameters     map[string]interface{} `json:"modelParameters,omitempty"`
	StartTime           time.Time              `json:"startTime,omitempty"`
	EndTime             time.Time              `json:"endTime,omitempty"`
	PromptName          string                 `json:"promptName,omitempty"`
	PromptVersion       string                 `json:"promptVersion,omitempty"`
}

func (g *Generation) End() error {
	if g.ID == "" {
		return errors.New("span id is not set")
	}
	g.EndTime = time.Now()
	err := g.eventManager.Enqueue(g.ID, GENERATION_UPDATE, g)
	return err
}

type Event struct {
	BasicObservation
	StartTime time.Time `json:"start_time"`
}

type Score struct {
	BasicObservation
	Value         int    `json:"value,omitempty,default:0"`
	ObservationId string `json:"observationId,omitempty"`
	Comment       string `json:"comment,omitempty"`
}
