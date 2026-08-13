# Coding Rules — Go + React (Wails)

These are the project's coding rules for implementation work. They apply in **coding mode** (see `knowledge.md`), after activating `/ponytail`.

Division of labor: **ponytail** governs *what* you build (the lazy-senior-dev ladder: YAGNI → reuse → stdlib → native → installed dep → one line → minimal). These rules govern *how* it fits this codebase. **Both apply.**

## File Size Limit (hard ceiling)

| Language / file type | Max lines per file |
|----------------------|--------------------|
| Go (`.go`) | **300** |
| TypeScript / TSX | **300** |
| CSS | **300** |

- The cap is a **ceiling, not a target** — prefer files of ~100–200 lines doing one thing. A file is a smell long before 300 lines.
- A file crossing 300 lines **must be split along a responsibility boundary** — never mechanically to hit a number (ponytail: fewest files possible).
  - Go: split per type / per concern within the package (e.g., `scanner.go`, `catalog.go`, `streamer.go`).
  - React: extract a component, custom hook, util, or constants file.
- **Exempt:** generated code, migrations, seed/fixture data, long constant/array tables, vendored code.

## Go

- `gofmt`/`goimports` clean; `go vet ./...` passes.
- Errors are explicit and wrapped (`fmt.Errorf("...: %w", err)`); never swallow errors; no panics outside fatal startup; `defer` for cleanup.
- Thread `context.Context` as the first parameter through anything that can block or cancel (blueprint: `context.WithCancel` for the `mariadb` subprocess).
- No goroutine leaks; use `errgroup` for fan-out; protect shared state with mutex/atomic (blueprint: atomic byte-progress counters).
- Stdlib first: `bufio`, `io.MultiReader`, `io.SectionReader`, `os/exec`, `database/sql` — add a dependency only when it does something stdlib can't.
- Package per responsibility; small exported API; idiomatic naming; short receivers.
- Tests: `_test.go` beside the code, table-driven, run with `go test ./...`.

## React / TypeScript (Wails frontend)

- TypeScript strict; no `any` without justification.
- Function components with typed props; custom hooks for shared logic; hooks at top level only.
- State local-first (`useState`/`useReducer`); add a global store only when state is genuinely shared across components.
- Large tables/lists use **virtual scrolling** (blueprint: >10k-table schema grid); throttle IPC progress events (~100–250ms ticker) to avoid flooding the bridge.
- Effect cleanup: remove listeners/subscriptions, cancel in-flight work on unmount.
- `memo`/`useMemo` only when measured; styling follows the existing convention, no new CSS framework without a reason.

## Verification (before finishing)

- Quality gates: `go build ./... && go vet ./... && go test ./...`; frontend typecheck/build via the project's scripts.
- `gitnexus_detect_changes()` before committing to confirm the blast radius matches your intent.

## Working from a Claim-Order Checklist (beads + MD)

When an issue comes from a claim-order MD file (`docs/plans/<plan>-claim-order.md`, produced in grill mode per `knowledge.md`), keep the issue and the checklist in sync through this loop:

### 1. Before starting — claim it

```bash
bd update <ID> --claim        # claim atomically; prevents double-work
```

Then update the checklist in the MD file — tick the box and note the claim:

```markdown
- [x] bd-XX — <slice title> — claimed — blocked by: bd-YY
```

### 2. Do the work

A normal coding task: ponytail ladder + the rules in this file. The checklist is not touched mid-work.

### 3. After finishing — close issue + update checklist

1. Close the issue: `bd close <ID>`
2. Update the checklist in the MD file — mark the item closed:

```markdown
- [x] bd-XX — <slice title> — ✅ closed — blocked by: bd-YY
```

### 4. Finalization — commit & push

Commit the code changes **together with the updated checklist MD**, then push:

```bash
git add <changed files> docs/plans/<plan>-claim-order.md
git commit -m "<ID>: <summary — closes issue + checklist>"
git push
```

An issue is only truly done when its code **and** its checklist state are committed and pushed — a stale checklist lets another agent re-claim the same issue. For the full session-completion ritual (`git pull --rebase`, status verification), follow AGENTS.md.
