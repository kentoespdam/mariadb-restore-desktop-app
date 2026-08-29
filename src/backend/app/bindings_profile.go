package app

import "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/profile"

// ListServerProfiles returns the redacted views.
func (a *App) ListServerProfiles() ([]profile.View, error) {
	return a.Profile.List()
}

// CreateServerProfile saves a new profile and returns its id.
func (a *App) CreateServerProfile(in profile.Input) (string, error) {
	return a.Profile.Create(in)
}

// UpdateServerProfile rewrites an existing profile (empty password keeps the old one).
func (a *App) UpdateServerProfile(in profile.Input) error {
	return a.Profile.Update(in)
}

// DeleteServerProfile removes a profile by id.
func (a *App) DeleteServerProfile(id string) error {
	return a.Profile.Delete(id)
}
