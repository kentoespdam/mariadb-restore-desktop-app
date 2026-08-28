// Launch wrapper for the Linux Wails binary.
//
// Strips any LD_LIBRARY_PATH entries pointing inside a snap mount
// (e.g. /snap/core20/...) before exec'ing the binary in the same directory.
// Snap core20 ships glibc 2.31's libpthread.so.0, which references
// __libc_pthread_init (GLIBC_PRIVATE) — a symbol the host libc.so.6
// (e.g. glibc 2.39 on Ubuntu 24.04) does not export — so the dynamic
// linker aborts the process with a "symbol lookup error" when the binary
// is launched from a snap-wrapped shell or IDE that injects the snap lib
// directory onto LD_LIBRARY_PATH.
//
// Build: CGO_ENABLED=0 go build -o build/launch-linux build/launch-linux.go
//
//	(static, so this wrapper is itself immune to a poisoned LD_LIBRARY_PATH)
//
// ponytail: this is a one-purpose wrapper, no config, no flags. The scrub
// list is "/snap/" — if/when a non-snap dist path also poisons (rare),
// extend the prefix list in one place.
package main

import (
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		os.Stderr.WriteString("usage: launch-linux <prog> [args...]\n")
		os.Exit(2)
	}
	prog := os.Args[1]

	env := os.Environ()
	clean := env[:0]
	for _, kv := range env {
		if strings.HasPrefix(kv, "LD_LIBRARY_PATH=") {
			val := strings.TrimPrefix(kv, "LD_LIBRARY_PATH=")
			paths := strings.Split(val, ":")
			kept := paths[:0]
			for _, p := range paths {
				if !strings.HasPrefix(p, "/snap/") {
					kept = append(kept, p)
				}
			}
			if len(kept) > 0 {
				clean = append(clean, "LD_LIBRARY_PATH="+strings.Join(kept, ":"))
			}
			continue
		}
		if strings.HasPrefix(kv, "LD_PRELOAD=") {
			val := strings.TrimPrefix(kv, "LD_PRELOAD=")
			paths := strings.Fields(val)
			kept := paths[:0]
			for _, p := range paths {
				if !strings.HasPrefix(p, "/snap/") {
					kept = append(kept, p)
				}
			}
			if len(kept) > 0 {
				clean = append(clean, "LD_PRELOAD="+strings.Join(kept, " "))
			}
			continue
		}
		clean = append(clean, kv)
	}

	cmd := exec.Command(prog, os.Args[2:]...)
	cmd.Env = clean
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		os.Exit(1)
	}
}
