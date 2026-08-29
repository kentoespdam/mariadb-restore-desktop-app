// Package recovery handles the "app.key missing but catalog present"
// state. It emits an event the frontend listens to, waits for a
// decision, and either wipes the catalog or returns a user-cancelled
// error.
package recovery

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/platform/events"
)

// Decision is the user's choice in the Smart Recovery modal.
type Decision string

const (
	DecisionCancel Decision = "cancel"
	DecisionReset  Decision = "reset"
)

// ErrUserCancelled is returned by HandleMissingKey when the user
// dismisses the modal without resetting.
var ErrUserCancelled = errors.New("recovery: user cancelled")

// ErrModalTimeout is returned when the frontend never answers.
var ErrModalTimeout = errors.New("recovery: modal timed out")

// ModalEvent is the payload the frontend receives on "recovery:show".
type ModalEvent struct {
	Reason string `json:"reason"`
}

// Modal abstracts the front-end modal so the service can be tested
// with a fake. The default implementation emits events on the bus
// and reads a one-shot response.
type Modal interface {
	// Show blocks until the user makes a decision or ctx is cancelled.
	// Returns ErrModalTimeout if no answer arrives in time.
	Show(ctx context.Context) (Decision, error)
}

// EventModal is the production implementation that talks to the Wails
// event bus.
type EventModal struct {
	Bus      events.Emitter
	Decision chan Decision
}

// Show emits the modal event and waits for a decision.
func (m *EventModal) Show(ctx context.Context) (Decision, error) {
	if m.Bus == nil {
		return "", fmt.Errorf("recovery: nil event bus")
	}
	if err := m.Bus.Emit(ctx, "recovery:show", ModalEvent{Reason: "missing_key"}); err != nil {
		return "", err
	}
	if m.Decision == nil {
		return "", fmt.Errorf("recovery: no decision channel")
	}
	select {
	case d := <-m.Decision:
		return d, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// fakeDecision lets tests drive a Modal synchronously.
type fakeDecision struct {
	mu sync.Mutex
	d  Decision
}

func (f *fakeDecision) Set(d Decision) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.d = d
}

func (f *fakeDecision) Show(_ context.Context) (Decision, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.d, nil
}
