// Seed: create a fresh app.key + catalog with one Server Profile
// pointing to the docker MariaDB. Run with: go run ./cmd/seed <dir>
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/app"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/profile"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: seed <dir>")
		os.Exit(2)
	}
	dir := os.Args[1]

	keyPath := crypto.KeyPath(dir)
	if _, err := crypto.GenerateKey(keyPath); err != nil {
		fmt.Fprintln(os.Stderr, "gen key:", err)
		os.Exit(1)
	}
	key, err := crypto.LoadKey(keyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load key:", err)
		os.Exit(1)
	}
	catPath := dir + "/catalog.sqlite"
	cat, err := catalog.Open(catPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open cat:", err)
		os.Exit(1)
	}
	defer cat.Close()

	// We don't need the Wails app to seed; the profile service
	// only needs the catalog + key. Build one in-process.
	profSvc := profile.New(cat, key)
	id, err := profSvc.Create(profile.Input{
		Name:     "docker-root",
		Host:     "127.0.0.1",
		Port:     3307,
		User:     "root",
		Password: "testpass",
		SSLMode:  "preferred",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "create profile:", err)
		os.Exit(1)
	}
	fmt.Printf("seeded profile id=%s in %s\n", id, dir)
	_ = app.New // keep the import alive in case future seeds need it
	_ = context.Background
}
