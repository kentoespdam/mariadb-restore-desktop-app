// Package events is a thin wrapper around the Wails runtime event bus.
// It has no business logic and never imports from features/.
package events

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Event is the wire-level payload delivered to subscribers.
type Event struct {
	Name    string
	Payload []byte
}

// Emitter is the contract both the real Wails bus and the test fake satisfy.
// The platform exposes a default Wails-backed implementation; tests inject
// a fake.
type Emitter interface {
	Emit(ctx context.Context, name string, payload any) error
}

// Default returns the Wails-backed emitter. The Wails ctx is looked
// up dynamically via GetWailsContext at Emit time so the emitter
// doesn't capture a stale context (e.g. context.Background passed at
// New before OnStartup ran). Call SetWailsContext once from OnStartup
// to register the live ctx.
func Default(_ context.Context) Emitter { return wailsEmitter{} }

type wailsEmitter struct{}

func (wailsEmitter) Emit(_ context.Context, name string, payload any) error {
	ctx := GetWailsContext()
	if ctx == nil {
		return fmt.Errorf("events: no wails ctx registered")
	}
	runtime.EventsEmit(ctx, name, payload)
	return nil
}

// Emit is a free-function shortcut for the default Wails bus. Pass a
// context captured from the OnStartup hook as the first arg.
func Emit(ctx context.Context, name string, payload any) error {
	return Default(ctx).Emit(ctx, name, payload)
}

// Subscribe is provided so features that need a synchronous stream can
// register a handler. Returns an unsubscribe function.
func Subscribe(name string, fn func(Event)) (func(), error) {
	// We use EventsOn — the unsubscribe is the returned cancel function.
	// NOTE: this requires a wails ctx; the caller must capture it via
	// the package-level SetWailsContext helper, or use SubscribeCtx
	// below when the ctx is in scope.
	if globalWailsCtx == nil {
		return func() {}, fmt.Errorf("events: no wails ctx registered")
	}
	return runtime.EventsOn(globalWailsCtx, name, func(payload ...any) {
		fn(Event{Name: name})
	}), nil
}

var globalWailsCtx context.Context

// SetWailsContext registers the active Wails context. Call once from
// OnStartup.
func SetWailsContext(ctx context.Context) { globalWailsCtx = ctx }

// GetWailsContext returns the Wails ctx registered via SetWailsContext.
// Returns nil if not yet set (e.g. OnStartup has not run).
func GetWailsContext() context.Context { return globalWailsCtx }
