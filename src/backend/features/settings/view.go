package settings

import "fmt"

// View is the projection the FE Settings screen renders. It bundles
// the on-disk Settings with the catalog / app.key paths from the
// Executable Scope (CONTEXT) so the screen has a single source.
type View struct {
	ExeDir         string `json:"exeDir"`
	CatalogPath    string `json:"catalogPath"`
	AppKeyPath     string `json:"appKeyPath"`
	MariadbPath    string `json:"mariadbPath"`
	MariadbDump    string `json:"mariadbDumpPath"`
	KeyBits        int    `json:"keyBits"`
}

// BuildView assembles a View from the service state plus the catalog
// + app.key paths the BE holds. The catalog/key paths are passed in
// so settings/ does not import core/crypto.
func (s *Service) BuildView(catalogPath, appKeyPath string) (View, error) {
	cur, err := s.Get()
	if err != nil {
		return View{}, fmt.Errorf("settings: build view: %w", err)
	}
	return View{
		ExeDir:      s.exeDir,
		CatalogPath: catalogPath,
		AppKeyPath:  appKeyPath,
		MariadbPath: cur.MariadbPath,
		MariadbDump: cur.MariadbDumpPath,
		KeyBits:     s.keyBits,
	}, nil
}

// ExeDir exposes the on-disk directory so the binding layer can
// surface it in the Settings View without callers poking into the
// package internals.
func (s *Service) ExeDir() string { return s.exeDir }