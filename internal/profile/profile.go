package profile

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"mariadb-restore-desktop-app/internal/key"
)

// Profile represents a server connection profile.
type Profile struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Username          string `json:"username"`
	EncryptedPassword []byte `json:"-"`
	Password          string `json:"password,omitempty"` // plaintext input only
	SSLMode           string `json:"sslMode"`
	SSLCA             string `json:"sslCa"`
	SSLCert           string `json:"sslCert"`
	SSLKey            string `json:"sslKey"`
	CreatedAt         string `json:"createdAt"`
}

// TestResult holds the outcome of a connection test.
type TestResult struct {
	OK      bool   `json:"ok"`
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListProfiles returns all profiles from the catalog.
func ListProfiles(db *sql.DB) ([]Profile, error) {
	rows, err := db.Query(
		"SELECT id, name, host, port, username, encrypted_password, ssl_mode, ssl_ca, ssl_cert, ssl_key, created_at FROM server_profiles ORDER BY name",
	)
	if err != nil {
		return nil, fmt.Errorf("query profiles: %w", err)
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Name, &p.Host, &p.Port, &p.Username, &p.EncryptedPassword, &p.SSLMode, &p.SSLCA, &p.SSLCert, &p.SSLKey, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// GetProfile returns a single profile by ID.
func GetProfile(db *sql.DB, id int64) (*Profile, error) {
	var p Profile
	err := db.QueryRow(
		"SELECT id, name, host, port, username, encrypted_password, ssl_mode, ssl_ca, ssl_cert, ssl_key, created_at FROM server_profiles WHERE id=?", id,
	).Scan(&p.ID, &p.Name, &p.Host, &p.Port, &p.Username, &p.EncryptedPassword, &p.SSLMode, &p.SSLCA, &p.SSLCert, &p.SSLKey, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return &p, nil
}

// CreateProfile inserts a new profile, encrypting the password with app.key.
func CreateProfile(db *sql.DB, k []byte, p *Profile) (int64, error) {
	if p.Port == 0 {
		p.Port = 3306
	}
	if p.SSLMode == "" {
		p.SSLMode = "disabled"
	}

	var encPass []byte
	if p.Password != "" {
		var err error
		encPass, err = key.Encrypt(k, []byte(p.Password))
		if err != nil {
			return 0, fmt.Errorf("encrypt password: %w", err)
		}
	}

	res, err := db.Exec(
		"INSERT INTO server_profiles (name, host, port, username, encrypted_password, ssl_mode, ssl_ca, ssl_cert, ssl_key) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.Host, p.Port, p.Username, encPass, p.SSLMode, p.SSLCA, p.SSLCert, p.SSLKey,
	)
	if err != nil {
		return 0, fmt.Errorf("insert profile: %w", err)
	}
	return res.LastInsertId()
}

// UpdateProfile updates an existing profile. If Password is non-empty, re-encrypts it.
func UpdateProfile(db *sql.DB, k []byte, p *Profile) error {
	if p.SSLMode == "" {
		p.SSLMode = "disabled"
	}

	if p.Password != "" {
		encPass, err := key.Encrypt(k, []byte(p.Password))
		if err != nil {
			return fmt.Errorf("encrypt password: %w", err)
		}
		_, err = db.Exec(
			"UPDATE server_profiles SET name=?, host=?, port=?, username=?, encrypted_password=?, ssl_mode=?, ssl_ca=?, ssl_cert=?, ssl_key=? WHERE id=?",
			p.Name, p.Host, p.Port, p.Username, encPass, p.SSLMode, p.SSLCA, p.SSLCert, p.SSLKey, p.ID,
		)
		if err != nil {
			return fmt.Errorf("update profile: %w", err)
		}
		return nil
	}

	_, err := db.Exec(
		"UPDATE server_profiles SET name=?, host=?, port=?, username=?, ssl_mode=?, ssl_ca=?, ssl_cert=?, ssl_key=? WHERE id=?",
		p.Name, p.Host, p.Port, p.Username, p.SSLMode, p.SSLCA, p.SSLCert, p.SSLKey, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

// DeleteProfile removes a profile by ID.
func DeleteProfile(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM server_profiles WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	return nil
}

// DecryptPassword decrypts the stored password using app.key.
func DecryptPassword(k, encrypted []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}
	plain, err := key.Decrypt(k, encrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt password: %w", err)
	}
	return string(plain), nil
}

// TestConnection attempts to connect to the MariaDB server using the default mariadb CLI.
func TestConnection(ctx context.Context, p *Profile, password string) TestResult {
	return TestConnectionWithBinary(ctx, "mariadb", p, password)
}

// TestConnectionWithBinary attempts to connect to the MariaDB server using the specified binary path.
func TestConnectionWithBinary(ctx context.Context, bin string, p *Profile, password string) TestResult {
	if bin == "" {
		bin = "mariadb"
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	args := []string{
		"-h", p.Host,
		"-P", fmt.Sprintf("%d", p.Port),
		"-u", p.Username,
		"--connect-timeout=5",
	}
	if password != "" {
		args = append(args, "-p"+password)
	}
	if p.SSLMode != "" && p.SSLMode != "disabled" {
		args = append(args, "--ssl-mode="+p.SSLMode)
		if p.SSLCA != "" {
			args = append(args, "--ssl-ca="+p.SSLCA)
		}
		if p.SSLCert != "" {
			args = append(args, "--ssl-cert="+p.SSLCert)
		}
		if p.SSLKey != "" {
			args = append(args, "--ssl-key="+p.SSLKey)
		}
	}
	args = append(args, "-e", "SELECT VERSION()")

	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return TestResult{OK: false, Error: msg}
	}

	version := strings.TrimSpace(string(out))
	return TestResult{OK: true, Version: version}
}
