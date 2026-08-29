package app

import (
	"context"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/recovery"
)

// HandleMissingKey is the Wails-callable Smart Recovery entry. The
// frontend invokes this when the boot probe reports app.key is gone
// but a catalog exists. The user picks Cancel or Reset in the modal;
// the result is returned as a string so the TS type is trivial.
func (a *App) HandleMissingKey() (string, error) {
	return a.Recovery.Handle(context.Background())
}

// RecoveryDecision is the inverse channel: the frontend calls this to
// send the user's choice back into the modal decision channel. The
// channel is buffered (size 1) so a decision sent while no modal is
// waiting is dropped silently.
func (a *App) RecoveryDecision(decision string) error {
	d := recovery.Decision(decision)
	if d != recovery.DecisionCancel && d != recovery.DecisionReset {
		return nil
	}
	select {
	case a.Recovery.Decision <- d:
	default:
	}
	return nil
}
