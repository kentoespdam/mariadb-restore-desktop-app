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

// Default returns the Wails-backed emitter. Call from a goroutine that has
// access to a Wails ctx (typically the OnStartup hook).
func Default(ctx context.Context) Emitter { return wailsEmitter{ctx: ctx} }

type wailsEmitter struct{ ctx context.Context }

func (w wailsEmitter) Emit(_ context.Context, name string, payload any) error {
	if w.ctx == nil {
		return fmt.Errorf("events: nil wails ctx")
	}
	runtime.EventsEmit(w.ctx, name, payload)
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
