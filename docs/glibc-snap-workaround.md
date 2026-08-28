# Workaround for the snap-core20 GLIBC symbol-lookup crash

## Symptom

Launching the Linux dev binary (`build/bin/mariadb-restore-desktop-app` or
`build/bin/mariadb-restore-desktop-app-dev-linux-amd64`) fails immediately with:

```
(mariadb-restore-desktop-app-dev-linux-amd64:NNNNN): Gtk-WARNING **: HH:MM:SS: Theme parsing error: gtk.css:NNN:NN: '-gtk-icon-size' is not a valid property name
mariadb-restore-desktop-app-dev-linux-amd64: symbol lookup error: /snap/core20/current/lib/x86_64-linux-gnu/libpthread.so.0: undefined symbol: __libc_pthread_init, version GLIBC_PRIVATE
```

The `Gtk-WARNING` line is cosmetic (host GTK theme has a non-fatal CSS quirk).
The actual failure is the second line.

## Cause

The launcher process has `/snap/core20/current/lib/x86_64-linux-gnu` on its
library search path (`LD_LIBRARY_PATH` or via `LD_PRELOAD`). Snap core20 ships
**glibc 2.31's `libpthread.so.0`**, which calls `__libc_pthread_init` — a
`GLIBC_PRIVATE` symbol that only exists in glibc 2.31's `libc.so.6`.

The Wails binary was built on the host (Ubuntu 24.04, glibc 2.39) and resolves
its own `libc.so.6` to the host's glibc 2.39. The dynamic linker ends up
loading the snap's 2.31 `libpthread.so.0` against the host's 2.39
`libc.so.6` → undefined symbol → process killed before `main()` runs.

Common source of the poisoned env: a snap-packaged terminal, IDE, or file
manager that prepends `/snap/core20/current/lib/x86_64-linux-gnu` to
`LD_LIBRARY_PATH` for its own sandbox.

## Fix (built in)

`build/launch-linux` is a tiny **statically-linked** Go binary that scrubs any
`/snap/...` entries from `LD_LIBRARY_PATH` and `LD_PRELOAD`, then `execvp`s
the real Wails binary with the cleaned env. It is immune to the poisoned
env itself because it doesn't link against any shared libs.

The `make build` (and `make build-linux`) targets build it automatically next
to the Wails binary.

## Usage

Instead of:

```bash
./build/bin/mariadb-restore-desktop-app
```

run:

```bash
./build/launch-linux ./build/bin/mariadb-restore-desktop-app
```

All CLI args are forwarded:

```bash
./build/launch-linux ./build/bin/mariadb-restore-desktop-app --flag
```

## Verifying the fix

```bash
build/launch-verify.sh
```

Reproduces the original crash (red) and confirms the launcher resolves it
(green). Requires `cc` and a `snap core20` install at
`/snap/core20/current/lib/x86_64-linux-gnu` to reproduce; otherwise skipped.

## What we skipped

- **Bundling a fixed `libc.so.6`/`libpthread.so.0` with the binary
  (AppImage-style).** Heavy: requires shipping ~10 MB of glibc + patchelf,
  and the WebKit2GTK transitive deps make the right subset non-obvious.
  Add when distribution to non-snap-wrapped hosts is required.
- **Building the Wails binary fully static.** Officially unsupported by
  Wails — WebKit2GTK + fontconfig + ICU + cairo have dynamic-loading
  requirements that break under `-extldflags -static`.
- **Auto-detecting the bug at startup and printing a friendlier error.**
  Catches it at `main()` instead of link time, but doesn't actually fix
  anything. Worth adding only if user reports persist after the launcher
  lands.
