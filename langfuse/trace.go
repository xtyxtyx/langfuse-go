package langfuse

import "errors"

type Trace struct {
	BasicObservation
	UserID    string   `json:"userId,omitempty"`
	SessionID string   `json:"sessionId,omitempty"`
	Version   string   `json:"version,omitempty"`
	Release   string   `json:"release,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Public    bool     `json:"public"`
}

func (o Trace) Update() error {
	if o.ID == "" {
		return errors.New("trace id is not set")
	}

	o.eventManager.Enqueue("", TRACE_CREATE, o)
	return nil
}

func (o Trace) Span(span *Span) (*Span, error) {
	if span == nil {
		span = &Span{}
	}

	if span.TraceID == "" {
		span.TraceID = o.ID
	}

	return o.BasicObservation.Span(span)
}

func (o Trace) Event(event *Event) (*Event, error) {
	if event == nil {
		event = &Event{}
	}

	if event.TraceID == "" {
		event.TraceID = o.ID
	}

	return o.BasicObservation.Event(event)
}

func (o Trace) Generation(generation *Generation) (*Generation, error) {
	if generation == nil {
		generation = &Generation{}
	}

	if generation.TraceID == "" {
		generation.TraceID = o.ID
	}

	return o.BasicObservation.Generation(generation)
}
