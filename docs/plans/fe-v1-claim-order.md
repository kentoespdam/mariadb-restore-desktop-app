# Claim Order — fe-v1

Created: 2026-08-29 · Source: grill session per CONTEXT.md + ADR-0007/0008

Per `knowledge.md` grill mode this plan ships only docs + beads issues. Implementation lives in the issues below; an agent in coding mode claims and closes them. Claim with `bd update <ID> --claim`. Keep the issue and this checklist in sync through the loop in `docs/coding-rules.md`.

## Foundation (no blockers)

- [x] bd-aci — FE-01 useWailsEvent subscription hook — ✅ closed
- [x] bd-94m — FE-02 API module layer + typed stubs — ✅ closed

## Screens & modal (after foundation)

- [x] bd-070 — FE-03 Smart Recovery Dialog (global) — ✅ closed
- [x] bd-que — FE-04 Dashboard screen — ✅ closed
- [x] bd-m34 — FE-05 Backup screen — ✅ closed
- [x] bd-8ks — FE-06 Restore entry + Full restore — ✅ closed
- [x] bd-33h — FE-07 Analyze + Object Selection grid (Partial Restore) — ✅ closed
- [x] bd-el0 — FE-08 Settings screen — ✅ closed

## Assembly

- [x] bd-6ar — FE-09 App routing expansion — ✅ closed
- [x] bd-8xe — FE-10 Tests for new screens — ✅ closed
- [ ] bd-a8h — FE-11 Verification + cleanup — blocked by: FE-10

## Notes for claimers

- FE-01 and FE-02 are intentionally unblocked so two agents can work in parallel.
- FE-07 is the only P0; everything else is P1/P2.
- Only one new dependency is added: `@tanstack/react-virtual@^3` (in FE-07). Do not pull in any other dep without re-opening ADR-0007/0008.
- API modules throw `Error("not implemented: <name>")` — real BE bindings land in a follow-up plan and will not change screen code.
- After closing any issue, mark its box ✅ below and commit the updated MD alongside the code in the same commit (per `docs/coding-rules.md`).
