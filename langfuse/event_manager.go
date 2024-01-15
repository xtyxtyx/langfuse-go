//go:generate moq -out mocks_test.go -pkg langfuse_test . EventManager

package langfuse

import (
	"context"
)

type EventManager interface {
	Enqueue(id string, eventType string, event interface{}) error
	Flush(ctxt context.Context)
}
