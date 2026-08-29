package catalog

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
)

// Profile is a server connection profile with encrypted sensitive fields.
type Profile struct {
	ID        string
	Name      string
	Host      string
	Port      int
	User      string
	Password  string
	SSLMode   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ErrProfileNotFound is returned by LoadProfile when no row matches.
var ErrProfileNotFound = errors.New("catalog: profile not found")

// SaveProfile inserts or updates a profile, encrypting the sensitive
// fields with key.
func (s *Store) SaveProfile(p *Profile, key []byte) error {
	if p.ID == "" {
		id, err := newID()
		if err != nil {
			return err
		}
		p.ID = id
	}
	host, err := crypto.Encrypt(key, []byte(p.Host))
	if err != nil {
		return err
	}
	user, err := crypto.Encrypt(key, []byte(p.User))
	if err != nil {
		return err
	}
	pass, err := crypto.Encrypt(key, []byte(p.Password))
	if err != nil {
		return err
	}
	if p.SSLMode == "" {
		p.SSLMode = "preferred"
	}
	now := time.Now().Unix()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = nowTime(now)
	}
	p.UpdatedAt = nowTime(now)

	_, err = s.db.Exec(`
        INSERT INTO profiles(id, name, host, port, user, password, ssl_mode, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
            name=excluded.name,
            host=excluded.host,
            port=excluded.port,
            user=excluded.user,
            password=excluded.password,
            ssl_mode=excluded.ssl_mode,
            updated_at=excluded.updated_at
    `, p.ID, p.Name, host, p.Port, user, pass, p.SSLMode, p.CreatedAt.Unix(), p.UpdatedAt.Unix())
	if err != nil {
		return fmt.Errorf("catalog: save profile: %w", err)
	}
	return nil
}

// LoadProfile reads and decrypts the profile named name.
func (s *Store) LoadProfile(name string, key []byte) (*Profile, error) {
	row := s.db.QueryRow(`
        SELECT id, name, host, port, user, password, ssl_mode, created_at, updated_at
        FROM profiles WHERE name = ?`, name)
	var (
		p                Profile
		host, user, pass []byte
		created, updated int64
	)
	if err := row.Scan(&p.ID, &p.Name, &host, &p.Port, &user, &pass, &p.SSLMode, &created, &updated); err != nil {
		if errors.Is(err, sqlErrNoRows()) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("catalog: load profile: %w", err)
	}
	var derr error
	if p.Host, derr = decryptString(key, host); derr != nil {
		return nil, derr
	}
	if p.User, derr = decryptString(key, user); derr != nil {
		return nil, derr
	}
	if p.Password, derr = decryptString(key, pass); derr != nil {
		return nil, derr
	}
	p.CreatedAt = nowTime(created)
	p.UpdatedAt = nowTime(updated)
	return &p, nil
}

// ListProfiles returns the names of all saved profiles (no decryption).
func (s *Store) ListProfiles() ([]string, error) {
	rows, err := s.db.Query(`SELECT name FROM profiles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("catalog: list: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// DeleteProfile removes a profile by name.
func (s *Store) DeleteProfile(name string) error {
	_, err := s.db.Exec(`DELETE FROM profiles WHERE name = ?`, name)
	return err
}

func decryptString(key, blob []byte) (string, error) {
	pt, err := crypto.Decrypt(key, blob)
	if err != nil {
		return "", fmt.Errorf("catalog: decrypt: %w", err)
	}
	return string(pt), nil
}

func newID() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("catalog: new uuid: %w", err)
	}
	return id.String(), nil
}

func nowTime(unix int64) time.Time {
	return time.Unix(unix, 0).UTC()
}
