package langfuse

import (
	"errors"
	"time"

	"github.com/segmentio/ksuid"
)

type Observation interface {
	Span(opts *Span) (*Span, error)
	Event(opts *Event) (*Event, error)
	Generation(opts *Generation) (*Generation, error)
	Score(opts *Score) (*Score, error)
	Update(opts Observation) (Observation, error)
}

type BasicObservation struct {
	ID            string                 `json:"id,omitempty"`
	Name          string                 `json:"name,omitempty"`
	TraceID       string                 `json:"traceId,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Level         string                 `json:"level,omitempty"`
	StatusMessage string                 `json:"statusMessage,omitempty"`
	Input         interface{}            `json:"input,omitempty"`
	Output        interface{}            `json:"output,omitempty"`
	Version       string                 `json:"version,omitempty"`
	ParentID      string                 `json:"parentObservationId,omitempty"`
	eventManager  EventManager
}

func (o BasicObservation) Span(span *Span) (*Span, error) {
	if span == nil {
		span = &Span{}
	}

	if span.ID == "" {
		span.ID = ksuid.New().String()
	}

	//only set the parent id if it is not set and the parent is not the trace
	if span.ParentID == "" && o.ID != o.TraceID && o.ID != span.TraceID {
		span.ParentID = o.ID
	}

	if span.TraceID == "" {
		span.TraceID = o.TraceID
	}

	if span.StartTime.IsZero() {
		span.StartTime = time.Now()
	}

	span.eventManager = o.eventManager
	err := o.eventManager.Enqueue("", SPAN_CREATE, span)

	return span, err
}

func (o BasicObservation) Event(opts *Event) (*Event, error) {
	if opts == nil {
		opts = &Event{}
	}

	if opts.ID == "" {
		opts.ID = ksuid.New().String()
	}

	//only set the parent id if it is not set and the parent is not the trace
	if opts.ParentID == "" && o.ID != o.TraceID && o.ID != opts.TraceID {
		opts.ParentID = o.ID
	}

	if opts.TraceID == "" {
		opts.TraceID = o.TraceID
	}

	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}

	opts.eventManager = o.eventManager
	o.eventManager.Enqueue("", EVENT_CREATE, opts)

	return opts, nil
}

func (o BasicObservation) Generation(generation *Generation) (*Generation, error) {
	if generation == nil {
		generation = &Generation{}
	}

	if generation.ID == "" {
		generation.ID = ksuid.New().String()
	}

	//only set the parent id if it is not set and the parent is not the trace
	if generation.ParentID == "" && o.ID != o.TraceID && o.ID != generation.TraceID {
		generation.ParentID = o.ID
	}

	if generation.TraceID == "" {
		generation.TraceID = o.TraceID
	}

	if generation.StartTime.IsZero() {
		generation.StartTime = time.Now()
	}

	generation.eventManager = o.eventManager
	err := o.eventManager.Enqueue("", GENERATION_CREATE, generation)

	return generation, err
}

func (o BasicObservation) Score(opts *Score) (*Score, error) {
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
	o.eventManager.Enqueue("", SCORE_CREATE, opts)

	return opts, nil
}

// // Update the observation with new values
// func (o BasicObservation) Update(opts Observation) (Observation, error) {
// 	err := o.eventManager.Enqueue("", OBSERVATION_UPDATE, opts)
// 	return opts, err
// }

type Span struct {
	BasicObservation
	StartTime time.Time  `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
}

func (s *Span) Update() error {
	if s.ID == "" {
		return errors.New("span id is not set")
	}

	s.eventManager.Enqueue("", SPAN_UPDATE, s)

	return nil
}

func (s *Span) End() error {
	if s.ID == "" {
		return errors.New("span id is not set")
	}
	now := time.Now()
	s.EndTime = &now
	s.eventManager.Enqueue("", SPAN_UPDATE, s)
	return nil
}

type Generation struct {
	BasicObservation
	CompletionStartTime time.Time              `json:"completionStartTime,omitempty"`
	Model               string                 `json:"model,omitempty"`
	Usage               map[string]interface{} `json:"usage,omitempty"`
	ModelParameters     map[string]interface{} `json:"modelParameters,omitempty"`
	StartTime           time.Time              `json:"startTime,omitempty"`
	EndTime             *time.Time             `json:"endTime,omitempty"`
	PromptName          string                 `json:"promptName,omitempty"`
	PromptVersion       string                 `json:"promptVersion,omitempty"`
}

func (g *Generation) Update() error {
	if g.ID == "" {
		return errors.New("generation id is not set")
	}

	g.eventManager.Enqueue("", GENERATION_UPDATE, g)
	return nil
}

func (g *Generation) End() error {
	if g.ID == "" {
		return errors.New("span id is not set")
	}
	now := time.Now()
	g.EndTime = &now
	err := g.eventManager.Enqueue("", GENERATION_UPDATE, g)
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
