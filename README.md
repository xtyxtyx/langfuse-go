# Langfuse Go SDK
## Installation
```bash
go get github.com/wepala/langfuse-go
```
Langchain documentation: https://docs.langfuse.com/langchain

## Usage

```go
package main

import (
	"github.com/wepala/langfuse-go/langfuse"
	"net/http"
	"context"
)

func main() {
	sdk := langfuse.New(&langfuse.Options{
		HttpClient: &http.Client{},
		SecretKey:  "secret-key",
		PublicKey:  "public-key",
	})
	ctxt := context.Background()
	trace := sdk.Trace(ctxt, &langfuse.Trace{
		ID:        "request-id",
		Name:      "trace-name",
		UserID:    "user-id",
		SessionID: "session-id",
	})
	span := trace.Span(ctxt, &langfuse.SpanOptions{
		Name: "span-name",
	})
	span.Generation(ctxt, &langfuse.GenerationOptions{
		GenerationId: "generation-id",
		Attributes: map[string]interface{}{
			"key": "value",
		},
	})
	span.End()
}
```

### Development 

#### Architecture
```mermaid
classDiagram
    Langfuse: - *http.Client httpClient
    Langfuse: - EventManager eventManager
    Langfuse: +Trace(Context ctxt, *Trace opts) *Trace
    Langfuse: +Flush()
    class Observation
    <<interface>> Observation
    class EventManager
    <<interface>> EventManager
    Observation: +Span(Context ctxt,*Span opts) *Span
    Observation: +Score(Context ctxt, *Score opts) *Score
    Observation: +Event(Context ctxt, *Event opts) *Event
    Observation: +Generation(Context ctxt, *Generation opts) *Generation
    Observation: +Score(Context ctxt, *Score opts) *Score
    
    BasicObservation: - EventManager eventManager
    BasicObservation: +string ID
    BasicObservation: +string TraceID
    BasicObservation: +string ParentObservationID
    BasicObservation: +string Name
    BasicObservation: map[string]interface{} Input
    BasicObservation: map[string]interface{} Output
    BasicObservation: map[string]interface{} Metadata
    BasicObservation: +string Version
    
    BasicObservation: +WithEventManager(EventManager eventManager) *Observation

    BasicEventManager: -*api.Client client
    BasicEventManager: +int maxBatchItems
    BasicEventManager: -[]*Queue queues
    BasicEventManager: Enqueue(interface{} event)
    BasicEventManager: ProcessEvents([] *Event)
    
    Queue: -[] *Event events
    Queue: +Add(event *Event) error
    
    Trace: +string ID
    Trace: +string Name
    Trace: map[string]interface{} Input
    Trace: map[string]interface{} Output
    Trace: map[string]interface{} Metadata
    Trace: +string Version
    Trace: +string Release
    Span: +time.Time StartTime
    Span: +time.Time EndTime
    Span: +End()
    Generation: +time.Time StartTime
    Generation: +time.Time EndTime
    Generation: +End()
    Event: +time.Time StartTime
    Score: +float64 Value
    Score: +string Comment
    Score: +string TraceID
    Score: +string ObservationID

    Langfuse --> Trace : creates
    Langfuse ..|> Observation : implements
    BasicEventManager ..|> EventManager : implements
    BasicEventManager o-- Queue : aggregates
    Trace ..* BasicObservation 
    Trace ..> EventManager : calls
    Span ..* BasicObservation
    Span ..> EventManager : calls
    Generation ..* BasicObservation
    Generation ..> EventManager : calls
    Event ..* BasicObservation
    Event ..> EventManager : calls
    Score ..* BasicObservation
    Score ..> EventManager : calls
    Observation --> Span : creates
    Observation --> Generation : creates
    Observation --> Event : creates
    Observation --> Score : creates
    BasicObservation *-- Span
    BasicObservation -- EventManager : has
    BasicObservation ..|> Observation : implements
    
    
```
