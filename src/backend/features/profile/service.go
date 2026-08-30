package profile

import (
	"fmt"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
)

// Service is the profile CRUD surface. It carries the key in memory so
// callers (backup, restore, settings) can ask for decrypted credentials
// without touching crypto themselves.
type Service struct {
	Store *catalog.Store
	Key   []byte
}

// New returns a service. The key is held by reference; rotate it via
// Reload if the app.key file is regenerated (e.g. Smart Recovery reset).
func New(store *catalog.Store, key []byte) *Service {
	return &Service{Store: store, Key: key}
}

// Create saves a new profile. Returns the assigned ID.
func (s *Service) Create(in Input) (string, error) {
	if in.Name == "" {
		return "", fmt.Errorf("profile: name required")
	}
	p := &catalog.Profile{
		Name:     in.Name,
		Host:     in.Host,
		Port:     in.Port,
		User:     in.User,
		Password: in.Password,
		SSLMode:  in.SSLMode,
	}
	if err := s.Store.SaveProfile(p, s.Key); err != nil {
		return "", err
	}
	return p.ID, nil
}

// Update rewrites an existing profile. An empty Password field keeps
// the current value (we re-encrypt the old one).
func (s *Service) Update(in Input) error {
	if in.ID == "" {
		return fmt.Errorf("profile: update requires ID")
	}
	// The catalog stores by name, so we need the *old* name to load
	// the existing record. We round-trip via List to find the row
	// that owns this ID.
	oldName, old, err := s.findByID(in.ID)
	if err != nil {
		return err
	}
	if in.Password == "" {
		in.Password = old.Password
	}
	// If the name changed, the catalog will fail with UNIQUE — wipe
	// the old row first.
	if oldName != in.Name {
		if err := s.Store.DeleteProfile(oldName); err != nil {
			return err
		}
	}
	p := &catalog.Profile{
		ID:       in.ID,
		Name:     in.Name,
		Host:     in.Host,
		Port:     in.Port,
		User:     in.User,
		Password: in.Password,
		SSLMode:  in.SSLMode,
	}
	return s.Store.SaveProfile(p, s.Key)
}

func (s *Service) findByID(id string) (string, *catalog.Profile, error) {
	names, err := s.Store.ListProfiles()
	if err != nil {
		return "", nil, err
	}
	for _, n := range names {
		p, err := s.Store.LoadProfile(n, s.Key)
		if err != nil {
			continue
		}
		if p.ID == id {
			return n, p, nil
		}
	}
	return "", nil, fmt.Errorf("profile: id %q not found", id)
}

// Creds is the decrypted credentials the backup/restore subprocesses
// need. Password is plaintext — this struct must never be returned
// to the FE; bindings layer must convert to a redacted view before
// crossing the Wails boundary.
type Creds struct {
	ID       string
	Name     string
	Host     string
	Port     int
	User     string
	Password string
	SSLMode  string
}

// CredentialsByID returns the decrypted profile. Used by backup and
// restore to feed their subprocess argv without ever exposing the
// password to the FE.
func (s *Service) CredentialsByID(id string) (Creds, error) {
	_, p, err := s.findByID(id)
	if err != nil {
		return Creds{}, err
	}
	return Creds{
		ID:       p.ID,
		Name:     p.Name,
		Host:     p.Host,
		Port:     p.Port,
		User:     p.User,
		Password: p.Password,
		SSLMode:  p.SSLMode,
	}, nil
}

// Delete removes a profile by id.
func (s *Service) Delete(id string) error {
	// catalog.Store keys by name; look up first
	all, err := s.Store.ListProfiles()
	if err != nil {
		return err
	}
	for _, n := range all {
		p, err := s.Store.LoadProfile(n, s.Key)
		if err != nil {
			continue
		}
		if p.ID == id {
			return s.Store.DeleteProfile(n)
		}
	}
	return fmt.Errorf("profile: id %q not found", id)
}

// List returns the redacted View of every profile.
func (s *Service) List() ([]View, error) {
	names, err := s.Store.ListProfiles()
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(names))
	for _, n := range names {
		p, err := s.Store.LoadProfile(n, s.Key)
		if err != nil {
			return nil, err
		}
		out = append(out, View{
			ID:      p.ID,
			Name:    p.Name,
			Host:    p.Host,
			Port:    p.Port,
			User:    p.User,
			HasPass: p.Password != "",
			SSLMode: p.SSLMode,
		})
	}
	return out, nil
}
