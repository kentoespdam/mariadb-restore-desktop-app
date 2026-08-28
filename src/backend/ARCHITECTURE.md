# Backend Architecture

Source code lives in `src/backend/`. This file is the import-direction contract.

## Layout

```
src/backend/
├── core/          # pure infra — no Wails import, testable with `go test ./core/...` alone
│   ├── scanner/   # Byte-Offset Scanner
│   ├── catalog/   # SQLite store + AES-GCM-encrypted credentials
│   ├── streamer/  # Virtual Streamer + Definer Stripper + ProgressReader
│   └── crypto/    # AES-GCM primitives + app.key lifecycle
├── platform/      # Wails-aware infra primitives — shared, no business policy
│   └── events/    # Wails event emitter (Emit/Subscribe)
├── features/      # user-facing workflows + modal/policy handlers
│   ├── backup/
│   ├── restore/
│   ├── profile/
│   ├── settings/
│   ├── distribution/
│   └── recovery/  # Smart Recovery modal + policy
└── app/           # assembly — wires core + platform + features into one App
```

## Import direction (hard rule)

```
core        ← platform        ← features        ← app
   ↑                              ↑                ↑
   └──────────────┬───────────────┘                │
                  └────────────────────────────────┘
```

- `core/*` may import nothing from `platform/`, `features/`, or `app/`.
- `platform/*` may import `core/*` only.
- `features/*` may import `core/*` and `platform/*`. Features do not import each other unless explicitly listed below.
- `app/*` may import everything.

### Allowed cross-feature imports (exceptions)

- `features/restore/` may import `features/profile/` (Restore needs a target Server Profile).
- `features/backup/` may import `features/profile/` (Backup needs a source Server Profile).
- `features/recovery/` may import `features/profile/` (Reset & Re-init wipes credentials).

No other cross-feature imports. If you find yourself reaching across, the missing piece is likely a new package in `core/` or `platform/`.

## Binding methods (`app/bindings_*.go`)

Wails binding methods are thin wrappers — delegate to a service in `features/*` or `core/*` in one line. Do not put business logic in binding methods. Tests for logic live next to the service (`features/<name>/service_test.go`), not next to bindings.

## File size

Hard ceiling: 300 lines per Go file (`docs/coding-rules.md`). Split per responsibility, not mechanically.

## Verification

```bash
go build ./... && go vet ./... && go test ./...
```

`go test ./core/...` must run without Wails runtime. `go test ./platform/...` and `go test ./features/...` may need Wails mocks.
