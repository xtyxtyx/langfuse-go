package langfuse

type Trace struct {
	BasicObservation
	UserID    string   `json:"userId,omitempty"`
	SessionID string   `json:"sessionId,omitempty"`
	Version   string   `json:"version,omitempty"`
	Release   string   `json:"release,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Public    bool     `json:"public"`
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
