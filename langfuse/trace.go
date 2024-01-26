package langfuse

type Trace struct {
	BasicObservation
	UserID    string   `json:"userId"`
	SessionID string   `json:"sessionId"`
	Version   string   `json:"version"`
	Release   string   `json:"release"`
	Tags      []string `json:"tags"`
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
