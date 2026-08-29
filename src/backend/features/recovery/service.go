package recovery

import (
	"context"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
)

// Service is the recovery entry point used by app/. It owns the
// catalog handle, the on-disk paths to wipe, and the modal
// implementation.
type Service struct {
	Cat    *catalog.Store
	Paths  Paths
	Modal  Modal
}

// Handle is the Wails-callable entry point. It blocks until the
// modal returns a decision (or the user dismisses the app).
// Returns "reset" on success, "cancelled" on ErrUserCancelled, or
// "" with an error.
func (s *Service) Handle(ctx context.Context) (string, error) {
	if err := HandleMissingKey(ctx, s.Modal, s.Cat, s.Paths); err != nil {
		if err == ErrUserCancelled {
			return "cancelled", nil
		}
		return "", err
	}
	return "reset", nil
}
