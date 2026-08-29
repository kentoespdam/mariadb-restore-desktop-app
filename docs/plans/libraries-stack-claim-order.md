# Claim Order — Libraries & Stack Pins

Created: 2026-08-28 · Source: grill session (`/grill-with-docs`)

Purpose: apply the library & package decisions from ADR-0007 to the repo. Adds **two new slices** to the existing `src-layout-bootstrap` plan — one to extend the Wave-0 scaffold with the pinned deps, one for lint/format tooling — and updates the existing Wave-0 issue with the pin list so no implementer is left guessing.

The full library table lives in [ADR-0007](../adr/0007-library-stack-pins.md). This file is just the work order.

Claim issues in order. Before starting any issue: `bd update <ID> --claim`.
Tick the box when claimed; mark ✅ when closed.

---

## Wave 0 — Scaffold + stack pins (run sequentially)

- [x] `mariadb-restore-desktop-app-h76` — ✅ closed — Initialize Wails v2 scaffold with `src/` layout, pinning the exact `go.mod` / `package.json` deps from ADR-0007 — can start immediately. **The first step.**
- [x] `mariadb-restore-desktop-app-o8a` — ✅ closed — Tooling config: `.golangci.yml` + `biome.json` + `.editorconfig` + `lefthook.yml` — can start immediately (parallel with h76). Closes before any feature slice starts so all code lands already-linted.

After Wave 0 closes, the existing Wave 1 slices from `src-layout-bootstrap-claim-order.md` can be claimed in order.

---

## Quick reference — new IDs

| ID | Title | Est. |
|---|---|---|
| `h76` (updated) | Wails v2 scaffold + library pins from ADR-0007 | 30m |
| `o8a` (new) | Tooling config: golangci-lint + Biome + editorconfig + lefthook | 45m |

**Total estimated (new work):** ~75 minutes on top of the existing 7.5h bootstrap slice.

---

## Verification at the end of this slice

- `go mod tidy && go build ./...` succeeds with **only** the pinned deps in `go.mod` (no surprise transitive bulking).
- `npm ls --depth=0` shows only the pinned frontend deps (Tailwind, shadcn/ui peer deps, RHF, Zod, TanStack Virtual, Vitest, Biome — nothing else).
- `golangci-lint run ./...` exits 0 on the scaffolded repo.
- `npx biome check src/frontend` exits 0.
- `wails dev` opens the default window on Ubuntu 24.04 (`-tags webkit2_41`) and on Windows 11.
- `wails build` produces a Linux AppImage shell + a Windows zip shell (full AppImage/zip packaging is the responsibility of the `Distribution: bundle mariadb binaries` issue from the v1 plan).

---

## Reference

- [ADR-0007 — Library & package stack for v1](../adr/0007-library-stack-pins.md)
- [src-layout-bootstrap-claim-order.md](./src-layout-bootstrap-claim-order.md) — existing 9-issue foundation plan; h76 is part of it
- [Wails Linux build tags (Ubuntu 24.04)](https://wails.io/docs/gettingstarted/building/)
