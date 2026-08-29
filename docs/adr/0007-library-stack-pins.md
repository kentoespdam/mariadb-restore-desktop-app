# 0007 — Library & package stack for v1 (Ubuntu 24.04 + Windows 11)

The project's blueprint already pinned the high-level stack (Go + Wails + React + SQLite + mariadb CLI). This ADR pins the specific libraries and versions so every feature issue can rely on a single source of truth instead of re-deciding per slice.

## Backend (Go)

| Concern | Choice | Reason |
|---|---|---|
| App framework | `github.com/wailsapp/wails/v2 v2.15.x` | v3 is beta; v2 is the current stable. Matches ADR-0005 (AppImage bundle). |
| SQLite driver | `modernc.org/sqlite` | Pure-Go, no CGO. Cross-compile Windows from Ubuntu is one `GOOS=windows GOARCH=amd64 go build`. Avoids needing TDM-GCC on Windows. ~25% slower than `mattn/go-sqlite3` but our catalog workload is small (one row per dump object, single-writer). |
| Encryption | stdlib `crypto/aes` + `crypto/cipher` (GCM, 256-bit) | No third-party crypto. `app.key` is a 32-byte random file beside the binary; sensitive SQLite columns hold `base64(nonce ‖ ciphertext)`. |
| UUIDs | `github.com/gofrs/uuid` | Pure Go, supports UUIDv4 for Server Profile IDs and Catalog row IDs. No CGO. |
| Concurrency | `golang.org/x/sync/errgroup` | Sequential per ADR-0006 but errgroup.WithLimit(1) keeps cancellation propagation. |
| Logging | stdlib `log/slog` | Since Go 1.21 the ecosystem has converged on slog; no third-party logger needed. |
| HTTP | stdlib `net/http` on demand | v1 has no external HTTP calls. If/when added, stdlib is enough. |
| Tests | stdlib `testing` + `github.com/stretchr/testify` | Table-driven tests with readable assertions; mockery avoided — use interfaces + hand-rolled fakes. |
| Lint | `golangci-lint` default linter set | `govet`, `staticcheck`, `errcheck`, `goimports`, `ineffassign`, `unused`, `misspell`. Single `.golangci.yml` at repo root. |

## Frontend (React + TypeScript)

| Concern | Choice | Reason |
|---|---|---|
| Framework | `react@^19` + `react-dom@^19` | Wails `react-ts` template default. |
| Build | Vite (via Wails template) | Standard 2026 setup. |
| Styling | `tailwindcss@^4` + `@tailwindcss/vite` | Utility-first, no runtime cost. |
| Components | `shadcn/ui` (Radix primitives + Tailwind) | Out-of-the-box dialogs/dropdowns; tree-shaken. Requires `@radix-ui/*` packages. |
| Utilities | `class-variance-authority`, `clsx`, `tailwind-merge`, `lucide-react` (icons) | shadcn/ui dependency chain. |
| Forms | `react-hook-form@^7` + `zod@^4` + `@hookform/resolvers` | Type-safe schema validation. |
| Virtual scroll | `@tanstack/react-virtual` | 2026 consensus pick over `react-window`/`react-virtualized`. |
| State | `useState` + `useReducer` + Context | Per coding-rules.md: YAGNI global store. |
| Lint + format | `biome` (single tool) | Replaces ESLint + Prettier with one Rust binary. |
| Tests | `vitest` + `@testing-library/react` + `jsdom` | Vite-native test runner. |
| Pre-commit | `lefthook` | Single binary; simpler than husky. |

## Build & packaging

| Concern | Choice | Reason |
|---|---|---|
| Linux build | `wails build -tags webkit2_41` | Ubuntu 24.04 ships WebKitGTK 4.1, not 4.0; required build tag per Wails docs. |
| Windows build | `wails build -webview2 embed` | No runtime install needed on user machines. |
| Bundled MariaDB | MariaDB 10.11 LTS | Matches the target in `CONTEXT.md`. |
| Linux package | AppImage via `linuxdeploy` + `appimagetool` | ADR-0005. |
| Windows package | Portable zip (ADR-0005) | No installer needed; matches Executable Scope. |
| License | MIT for our code + MariaDB GPL notice | ADR-0005 consequence. |
| Versioning | SemVer via git tags + `-ldflags` | Wails official pattern. |
| CI | deferred to post-v1 | Per grill decision; not worth the setup for a 30h v1 plan. |

## Considered and rejected

- **`mattn/go-sqlite3`** — CGO; breaks Windows+Linux cross-compile.
- **`gosqlite.org/sqlite`** — newer and promising, but only ~1 year old; smaller community than `modernc.org/sqlite`.
- **SQLCipher** — whole-DB encryption is overkill; the blueprint only requires credentials encrypted, not metadata. Column-level AES-GCM with `app.key` is enough.
- **`rs/zerolog` / `uber-go/zap`** — log/slog is the stdlib consensus; no perf reason for v1's small log volume.
- **`react-window` / `react-virtualized`** — TanStack Virtual supersedes both in the 2026 community.
- **ESLint + Prettier** — Biome replaces both with one config.
- **CI (GitHub Actions, etc.)** — deferred; local builds only for v1.
- **Prometheus / OpenTelemetry** — YAGNI; slog covers debugging.
- **Registry-based mariadb discovery on Windows** — beside-binary + PATH + Settings override covers the common cases; no Registry dependency.
