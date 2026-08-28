# Portable MariaDB Backup & Restore

A cross-platform desktop application for backing up and restoring MariaDB databases with selective object-level restore capabilities. Built with [Wails v2](https://wails.io/) (Go + React/TypeScript).

![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-blue)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)
![License](https://img.shields.io/badge/license-MIT-green)

---

## Features

- **Portable** — single binary, runs from any directory, stores config next to the executable
- **Encrypted catalog** — server profiles and credentials stored in an encrypted local SQLite database
- **Smart Recovery** — automatic key recovery when catalog exists but encryption key is missing
- **Selective restore** — scan dump files, browse objects by database/table, restore individual objects
- **Multi-server profiles** — manage multiple MariaDB server connections
- **Binary auto-discovery** — automatically finds `mariadb` and `mariadb-dump` in `PATH`
- **Cross-platform** — Windows, Linux, and macOS support

---

## Tech Stack

| Layer    | Technology                    |
|----------|-------------------------------|
| Backend  | Go 1.25, Wails v2, SQLite3   |
| Frontend | React 19, TypeScript 5, Vite 7 |
| UI       | Custom CSS (no UI framework)  |

---

## Prerequisites

| Tool       | Version  | Purpose                     | Install                                      |
|------------|----------|-----------------------------|----------------------------------------------|
| Go         | ≥ 1.25   | Backend compiler            | [go.dev/dl](https://go.dev/dl/)              |
| Node.js    | ≥ 18     | Frontend build              | [nodejs.org](https://nodejs.org)             |
| npm        | ≥ 9      | Package manager             | Comes with Node.js                           |
| Wails CLI  | ≥ 2.15   | Desktop app build tool      | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` (binary lands in `$(go env GOPATH)/bin` — ensure it is on `PATH`) |
| MariaDB client | any | Backup/restore operations  | [mariadb.org](https://mariadb.org/download/) |

### System Dependencies

Wails v2 (≤ 2.15) hardcodes `webkit2gtk-4.0` in its cgo/pkg-config directives, which is no longer shipped by current Linux distributions. Newer distros only provide `webkit2gtk-4.1`, so the build also needs the `webkit2_41` build tag — `make` and the `Dockerfile` detect this automatically and apply it. When invoking `wails` directly, pass `-tags webkit2_41` yourself (see [Troubleshooting](#linux-build-fails-with-webkit2gtk-not-found) below).

**Linux (Debian/Ubuntu ≥ 22.04):**
```bash
sudo apt install -y libgtk-3-dev libwebkit2gtk-4.1-dev build-essential pkg-config
```
On Ubuntu 22.04 / Debian 12, replace `libwebkit2gtk-4.1-dev` with `libwebkit2gtk-4.0-dev`.

**Linux (Fedora/RHEL):**
```bash
sudo dnf install -y gtk3-devel webkit2gtk4.1-devel gcc pkg-config
```
On RHEL 9 / older Fedora, replace `webkit2gtk4.1-devel` with `webkit2gtk4.0-devel`.

**Windows:** No extra system dependencies (WebView2 is bundled).

**macOS:** Install Xcode Command Line Tools:
```bash
xcode-select --install
```

---

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/your-username/mariadb-restore-desktop-app.git
cd mariadb-restore-desktop-app

# 2. Install dependencies
make install

# 3. Run in development mode (with hot-reload)
make dev
```

---

## Build

### Using Make (Linux / macOS / Windows with Git Bash or WSL)

```bash
# Show all available commands
make help

# Install all dependencies
make install

# Run in development mode (hot-reload)
make dev

# Build for current platform
make build

# Cross-compile
make build-windows     # Windows amd64
make build-linux       # Linux amd64
make build-darwin      # macOS amd64 + arm64
make build-all         # All platforms

# Windows installer (NSIS)
make installer         # Windows amd64 installer
make installer-arm64   # Windows arm64 installer
make installer-all     # Both amd64 + arm64 installers
```

### Using Wails CLI directly

```bash
# Development mode
wails dev

# Production build (current platform)
wails build

# Cross-compile
wails build -platform windows/amd64
wails build -platform linux/amd64
wails build -platform darwin/arm64

# Windows installer (NSIS) — requires NSIS on PATH or bundled with Wails
wails build -platform windows/amd64 -nsis
```

### On Windows (without Make)

If you don't have `make` installed, you can run the Wails CLI commands directly, or install Make via:

```bash
# Via Chocolatey
choco install make

# Via Scoop
scoop install make

# Via MSYS2
pacman -S make
```

Alternatively, run the commands manually:

```powershell
# PowerShell
wails dev                  # Development mode
wails build                # Production build
```

Build output will be in `build/bin/`.

### Using Docker (Linux build)

Build a Linux binary inside a container — no local Go/Wails installation needed:

```bash
# Build the Docker image (includes frontend + backend compilation)
make docker-build

# Run the app with host display (X11)
make docker-run

# Or run manually
docker run --rm -it \
    -e DISPLAY=$DISPLAY \
    -v /tmp/.X11-unix:/tmp/.X11-unix \
    mariadb-restore-desktop-app:latest

# Open a shell inside the container for debugging
docker run --rm -it --entrypoint /bin/bash \
    mariadb-restore-desktop-app:latest
```

> **Note:** The Docker image requires a display server (X11/Wayland) to run the GUI.
> On Linux, pass `-e DISPLAY=$DISPLAY -v /tmp/.X11-unix:/tmp/.X11-unix`.
> On macOS, install [XQuartz](https://www.xquartz.org/) and enable TCP connections.

---

## Project Structure

```
mariadb-restore-desktop-app/
├── main.go                 # Wails entrypoint
├── app.go                  # App struct — all methods exposed to frontend
├── app_test.go             # Unit tests
├── go.mod / go.sum         # Go module files
├── wails.json              # Wails project config
├── Makefile                # Cross-platform build commands
├── Dockerfile              # Multi-stage Docker build (Linux)
├── .dockerignore           # Docker build context exclusions
│
├── internal/               # Go backend packages
│   ├── catalog/            #   Encrypted SQLite catalog for dump metadata
│   ├── key/                #   Encryption key generation & storage
│   ├── profile/            #   Server profile CRUD + connection testing
│   ├── scanner/            #   SQL dump file parser (object extraction)
│   └── settings/           #   Binary path discovery & persistence
│
├── frontend/               # React + TypeScript + Vite
│   ├── package.json
│   ├── vite.config.ts
│   ├── src/
│   │   ├── App.tsx         #   Root component
│   │   ├── components/     #   Reusable UI components
│   │   └── panels/         #   Panel views (servers, restore, settings)
│   └── wailsjs/            #   Auto-generated Wails bindings
│
└── build/                  # Build assets
    ├── appicon.png
    ├── windows/            #   Windows-specific (icon, manifest, installer)
    └── darwin/             #   macOS-specific (Info.plist)
```

---

## How It Works

### Application Lifecycle

1. **Startup** — App checks for encryption key (`key.dat`) and catalog database (`catalog.db`) next to the executable
2. **Fresh install** — Generates a new encryption key, creates the encrypted catalog
3. **Smart Recovery** — If catalog exists but key is missing, prompts to recover (use same key from previous install) or reset
4. **Ready** — Loads existing key and opens catalog

### Restore Workflow

1. **Select dump file** — Choose a `.sql` dump file to analyze
2. **Scan** — Parser extracts databases, tables, views, procedures, etc. with byte offsets
3. **Select objects** — Browse and check individual objects to restore
4. **Restore** — Executes selected objects against the target server

### Security

- Server passwords are encrypted with AES using a local key file
- The key file is stored alongside the binary (portable, not system-wide)
- Catalog data is stored in SQLite with SQLCipher encryption

---

## Testing

```bash
make test            # Run all tests
make test-verbose    # Verbose output
make test-race       # With race detector
```

Or directly with Go:

```bash
go test ./... -v
```

---

## Linting

```bash
make lint            # go vet + staticcheck
make vet             # go vet only
```

---

## Cleanup

```bash
make clean           # Remove build artifacts and frontend dist
```

---

## Docker

```bash
make docker-build    # Build Docker image (Linux amd64)
make docker-run      # Run with host X11 display
make docker-shell    # Open shell inside container
```

---

## Troubleshooting

### "Wails not found" / `make: wails: No such file or directory`
The Wails CLI was not installed or is not on `PATH`. Install it and ensure the install location is reachable from your shell:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
export PATH="$(go env GOPATH)/bin:$PATH"   # add to your ~/.bashrc / ~/.zshrc
wails version                                # sanity check
```

### `sh: 1: tsc: not found` during `make dev` / `wails dev`
**Symptom:** `wails dev` runs `npm run build` → `tsc && vite build` and dies with `tsc: not found`, even though `cd frontend && ls node_modules` looks populated.
**Root cause:** `npm install` skipped `devDependencies`. This happens when the shell sets `NODE_ENV=production` **and/or** npm is configured to omit dev deps (`npm config get omit` returns `dev`). In that state only the 4 runtime `dependencies` are installed, so `typescript`, `vite`, `@vitejs/plugin-react`, and the `@types/*` packages are missing — and `node_modules/.bin/tsc` does not exist.
**Fix:** the project already forces `npm install --include=dev` in `wails.json` (`frontend:install`) and the `Makefile` (`frontend-install`). If you still hit this, run install explicitly from a clean state:
```bash
rm -rf frontend/node_modules frontend/package-lock.json
cd frontend && npm install --include=dev && cd ..
make dev
```
To stop hitting it in this environment, unset `NODE_ENV` or set `npm config set omit ""` (project-level via `frontend/.npmrc` if only this repo should be affected).

### Linux build fails with "webkit2gtk not found"

Wails v2 (≤ 2.15) hardcodes `webkit2gtk-4.0`. If your distro (Ubuntu 24.04+, Fedora 40+, Linux Mint 22+) only ships `webkit2gtk-4.1`, install the matching `-dev` package and pass the `webkit2_41` build tag:

```bash
# Debian/Ubuntu
sudo apt install libwebkit2gtk-4.1-dev

# Fedora
sudo dnf install webkit2gtk4.1-devel

# Then build with the tag (Makefile and Dockerfile do this automatically)
wails build -tags webkit2_41
# or:
wails dev -tags webkit2_41
```

Reference: [Wails — Installation](https://wails.io/docs/gettingstarted/installation/) and [Wails — Compiling your Project](https://wails.io/docs/gettingstarted/building/).

### WebView2 not found on Windows
Download and install [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/).

### Frontend build fails
```bash
cd frontend && rm -rf node_modules && npm install && npm run build
```

---

## Author

**Bagus Sudrajat** — [kentoes.pdam@gmail.com](mailto:kentoes.pdam@gmail.com)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
