# Claim Order — v1 Portable MariaDB Backup & Restore

Created: 2026-08-13 · Source: grill session (blueprint.md + prd.md + CONTEXT.md)

Claim issues in order. Before starting any issue: `bd update <ID> --claim`.
Tick the box when claimed; mark ✅ when closed.

---

## Wave 1 — Must be first (blocks everything)

- [ ] `mariadb-restore-desktop-app-gco` — Bootstrap: Wails v2 app shell, SQLite+AES-GCM, Smart Recovery, sidebar scaffold — **can start immediately**

---

## Wave 2 — Parallel after Wave 1

All four can be claimed simultaneously by different agents once `gco` is closed.

- [ ] `mariadb-restore-desktop-app-yiq` — Server Profile CRUD with SSL config — blocked by: `gco`
- [ ] `mariadb-restore-desktop-app-q45` — Settings page: mariadb binary path config and auto-discovery — blocked by: `gco`
- [ ] `mariadb-restore-desktop-app-z3o` — Distribution: bundle mariadb binaries, Windows zip, Linux AppImage — blocked by: `gco`
- [ ] `mariadb-restore-desktop-app-pzn` — Byte-Offset Scanner: single-pass catalog population with Analyze UX — blocked by: `gco`

---

## Wave 3 — Parallel after Wave 2 (yiq + q45 must both be closed)

`zej` and `v29` can be claimed simultaneously.

- [ ] `mariadb-restore-desktop-app-zej` — Full Restore pipeline: stream file to mariadb CLI with transformers and progress — blocked by: `yiq`, `q45`
- [ ] `mariadb-restore-desktop-app-v29` — Backup workflow: multi-DB sequential mariadb-dump with progress — blocked by: `yiq`, `q45`

---

## Wave 4 — Last (pzn + zej must both be closed)

- [ ] `mariadb-restore-desktop-app-dbf` — Partial Restore pipeline: Analyze, checklist UI, Virtual Streamer, multi-DB sequential — blocked by: `pzn`, `zej`

---

## Quick reference — all IDs

| ID | Title | Est. |
|---|---|---|
| `gco` | Bootstrap: Wails v2 app shell + SQLite + Smart Recovery | 4h |
| `yiq` | Server Profile CRUD with SSL config | 3h |
| `q45` | Settings page + binary auto-discovery | 2h |
| `z3o` | Distribution: bundle binaries + AppImage | 3h |
| `pzn` | Byte-Offset Scanner + Analyze UX | 4h |
| `zej` | Full Restore pipeline | 5h |
| `v29` | Backup workflow | 4h |
| `dbf` | Partial Restore pipeline | 5h |

**Total estimated: ~30h**

---

## Key architectural decisions (ADRs)

- [ADR-0001](../adr/0001-backup-logical-not-physical.md) — Backup is logical (`mariadb-dump`), not physical
- [ADR-0002](../adr/0002-two-path-restore.md) — Restore has two paths: Full (no scan) and Partial (Analyze first)
- [ADR-0003](../adr/0003-drop-database-handling.md) — DROP DATABASE stripped in Partial, passed through in Full
- [ADR-0004](../adr/0004-single-slot-catalog.md) — Single-slot catalog; always fresh scan on new file
- [ADR-0005](../adr/0005-bundle-mariadb-binaries-appimage.md) — Bundle mariadb binaries; Linux as AppImage
- [ADR-0006](../adr/0006-multi-db-sequential-subprocess.md) — Multi-DB Partial Restore: sequential subprocess per database
