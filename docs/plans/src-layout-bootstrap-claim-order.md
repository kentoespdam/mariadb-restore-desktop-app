# Claim Order — src/ Layout Bootstrap

Created: 2026-08-28 · Source: grill session (`/grill-with-docs`)

Purpose: bootstrap the new `src/` layout (backend in `src/backend/`, frontend in `src/frontend/`), with `main.go`/`wails.json`/`go.mod` kept at repo root. Architectural contract: `src/backend/ARCHITECTURE.md`.

Claim issues in order. Before starting any issue: `bd update <ID> --claim`.
Tick the box when claimed; mark ✅ when closed.

---

## Wave 1 — Foundation (run in parallel, no blockers)

Three independent starter slices. Each can be claimed by a different agent.

- [ ] `mariadb-restore-desktop-app-72u` — Implement `core/crypto/` — AES-GCM primitives + `app.key` lifecycle — can start immediately
- [ ] `mariadb-restore-desktop-app-hgu` — Implement `core/scanner/` — single-pass Byte-Offset Scanner — can start immediately
- [ ] `mariadb-restore-desktop-app-xih` — Implement `platform/events/` — Wails event emitter primitive — can start immediately

After Wave 1 closes, the next three can be claimed:

- [ ] `mariadb-restore-desktop-app-efw` — Implement `core/catalog/` — SQLite store with AES-GCM credentials — blocked by: `72u`
- [ ] `mariadb-restore-desktop-app-ub9` — Implement `core/streamer/` — Virtual Streamer + Definer Stripper + ProgressReader — blocked by: `hgu`

---

## Wave 2 — Feature layer (after Wave 1 closes)

- [ ] `mariadb-restore-desktop-app-b64` — Implement `features/recovery/` — Smart Recovery modal + Reset policy — blocked by: `72u`, `efw`, `xih`
- [ ] `mariadb-restore-desktop-app-qbk` — Implement `features/profile/` — Server Profile CRUD — blocked by: `efw`

---

## Wave 3 — Assembly (last)

- [ ] `mariadb-restore-desktop-app-ap9` — Implement `app/` — assembly + Wails App constructor + binding stubs — blocked by: `b64`, `qbk`

---

## Wave 0 (separate concern — do this first to scaffold the project tree)

- [ ] `mariadb-restore-desktop-app-h76` — Initialize Wails v2 scaffold with `src/` layout — can start immediately. **This is the literal first step**; everything else assumes `src/backend/` and `src/frontend/` exist with the empty structure from `src/backend/ARCHITECTURE.md`. Run `wails init`, move `frontend/` → `src/frontend/`, edit `wails.json` paths, verify `wails dev` opens an empty window.

---

## Quick reference — all IDs

| ID | Title | Est. |
|---|---|---|
| `h76` | Initialize Wails v2 scaffold with `src/` layout | 30m |
| `72u` | `core/crypto/` — AES-GCM + `app.key` | 45m |
| `hgu` | `core/scanner/` — Byte-Offset Scanner | 60m |
| `xih` | `platform/events/` — Wails event emitter | 30m |
| `efw` | `core/catalog/` — SQLite + encrypted credentials | 60m |
| `ub9` | `core/streamer/` — Virtual Streamer + Definer Stripper | 90m |
| `b64` | `features/recovery/` — Smart Recovery modal | 60m |
| `qbk` | `features/profile/` — Server Profile CRUD | 45m |
| `ap9` | `app/` — assembly + binding stubs | 45m |

**Total estimated: ~7.5h** for the bootstrap slice (foundation + recovery + profile + app). Backup/Restore/Settings/Distribution features from the v1 plan are **out of scope** for this layout bootstrap — they are still tracked in the v1 plan (`docs/plans/v1-claim-order.md`) and will be re-issued against the new layout in a follow-up grill session.

---

## Layout summary

```
repo-root/
├── main.go             # thin entry: embed + wails.Run
├── wails.json          # frontend:install/build point at src/frontend
├── go.mod
├── src/
│   ├── backend/
│   │   ├── ARCHITECTURE.md
│   │   ├── core/{scanner,catalog,streamer,crypto}/
│   │   ├── platform/events/
│   │   ├── features/{backup,restore,profile,settings,distribution,recovery}/
│   │   └── app/
│   └── frontend/
│       ├── pages/{Dashboard,Backup,Restore,Profiles,Settings}/
│       ├── modals/
│       ├── components/
│       ├── hooks/
│       ├── app.tsx
│       └── main.tsx
└── (docs/, CONTEXT.md, AGENTS.md, CLAUDE.md, knowledge.md, beads, graphify — unchanged)
```

Import direction contract: `src/backend/ARCHITECTURE.md`.

---

## Frontend layout note

Frontend follows option (B) from Q10 of the grill session: organized by **what user sees** (pages + modals), not by mirroring backend features. The backend ↔ frontend correspondence is via Wails bindings, not folder structure. Junior devs finding "where does the Restore UI live" should look in `src/frontend/pages/Restore/`, not `src/frontend/features/restore/`.
