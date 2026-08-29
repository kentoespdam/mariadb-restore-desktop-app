// Package profile exposes the Server Profile CRUD service to the
// Wails binding layer. It is the single place that holds the app key
// in memory; password bytes are never returned to callers in
// plaintext.
package profile

// View is the projection of a profile the UI is allowed to see. It
// intentionally omits the password — the Wails binding surfaces a
// redacted form so the frontend never holds the cleartext.
type View struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	User    string `json:"user"`
	HasPass bool   `json:"hasPassword"`
	SSLMode string `json:"sslMode"`
}

// Input is what callers send for create / update. Empty Password
// keeps the existing one on update.
type Input struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"sslMode"`
}
