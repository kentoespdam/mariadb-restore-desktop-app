package events

import (
	"context"
	"errors"
	"testing"
)

type fakeEmitter struct {
	lastName    string
	lastPayload any
	err         error
}

func (f *fakeEmitter) Emit(_ context.Context, name string, payload any) error {
	f.lastName = name
	f.lastPayload = payload
	return f.err
}

func TestDefaultEmitterIsInterface(t *testing.T) {
	// Default needs a real wails ctx; we only test the contract via the
	// fake here, since instantiating a wails ctx requires a webview
	// process that does not exist in `go test`.
	var e Emitter = &fakeEmitter{}
	if err := e.Emit(context.Background(), "x", 1); err != nil {
		t.Fatal(err)
	}
}

func TestFakeEmitterRecordsPayload(t *testing.T) {
	f := &fakeEmitter{err: errors.New("boom")}
	var e Emitter = f
	if err := e.Emit(context.Background(), "progress", 42); err == nil {
		t.Fatal("want error")
	}
	if f.lastName != "progress" || f.lastPayload != 42 {
		t.Fatalf("recorded = %+v", f)
	}
}
