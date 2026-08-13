# Knowledge — Agent Operating Manual

This file defines **how agents work in this project**. Read it fully before starting any task, detect your mode, then follow that mode's rules strictly.

## Project Snapshot

- **Stack:** Go backend + Wails (React + TypeScript frontend) — portable MariaDB restore & backup desktop tool. See `docs/blueprint.md` and `docs/prd.md` for the agreed architecture (byte-offset scanner → SQLite catalog → virtual streamer → `mariadb` CLI subprocess).
- **Issue tracker:** beads (`bd`) — see AGENTS.md for the full workflow, session-completion protocol, and the "ALWAYS use non-interactive flags" shell rules.
- **Code intelligence:** graphify + gitnexus (see below). Output dirs: `graphify-out/`, `.gitnexus/`.

## Modes

Two modes of working. Detect yours, then follow its rules strictly.

| Mode | Trigger | May edit files? | Primary output |
|------|---------|-----------------|----------------|
| **Grill (planning)** | `/grill-me` or `/grill-with-docs` | **NO** | Beads issues with implementation plans + claim-order checklist MD |
| **Coding (implementation)** | Any task that writes/edits/refactors files | **YES** | Code, following the Coding Rules below |

If a coding task arrives without an explicit mode, you are in **coding mode**. Grill mode exists only while `/grill-me` or `/grill-with-docs` is active.

---

## Understanding the Code (all modes)

Use code-intelligence tools **first**; fall back to raw shell only when they don't cover the need.

### 1. graphify — codebase graph

Graph + report live in `graphify-out/` (`graph.json`, `graph.html`, `GRAPH_REPORT.md`). If missing or stale:

```bash
graphify update .        # incremental re-extraction — local AST, no LLM, no API keys
```

**Extraction is local and model-free** — the agent runs the CLI itself; no GEMINI or other model / API key. Full runbook: [Graph Update Process](#graph-update-process-on-demand).

Commands (default graph: `graphify-out/graph.json`):

```bash
graphify query "<question>"        # answer a question via BFS over the graph
graphify explain "<Symbol>"        # plain-language explanation of a node + its neighbors
graphify path "<A>" "<B>"          # dependency path between two symbols
graphify affected "<Symbol>"       # reverse traversal: what breaks if X changes
graphify god-nodes                 # most-connected nodes (architectural hubs)
```

**Full command reference** (quick recipes, exports, hooks, MCP server, model-free notes): `docs/graphify-commands.md`.

### 2. gitnexus — call-graph intelligence

- MCP tools (if available): `gitnexus_query({query})`, `gitnexus_context({name})`, `gitnexus_impact({target, direction})`, `gitnexus_detect_changes()`, `gitnexus_rename`.
- Resources: `gitnexus://repo/<repo>/context` · `/clusters` · `/processes` · `/process/<name>`.
- If the index is stale: `npx gitnexus analyze`.
- **Before editing any symbol you MUST run impact analysis** on it, and `gitnexus_detect_changes()` before committing — see the "Never Do" list in AGENTS.md.

### 3. Fallback — cat / grep / ls

Use plain shell when the graph tools can't answer: index missing/stale, exact-string searches, config/tooling files, brand-new code not yet indexed, or when you need raw source before making changes. Before falling back, check `docs/graphify-commands.md` — the graph can usually answer architecture/flow questions without shell.

```bash
cat <file>                       # read a file
grep -rn "<pattern>" .           # exact string search (use rg if available)
ls -R <dir>                      # discover structure
find . -name '*.go' -o -name '*.tsx'   # locate files by type
```

---

## Graph Update Process (on-demand)

When the user says **"update graph" / "refresh the graph"** (or asks to re-index the codebase), run **both**:

```bash
graphify update .        # incremental re-extraction — local AST, no LLM, no API keys
npx gitnexus analyze     # refresh the gitnexus index
```

`/graphify . --update` is the same graphify operation when invoked as the assistant skill; the agent runs the CLI equivalent directly (`graphify update .`). No model keys are involved in either step.

### How extraction works (researched from graphify source, v0.9.41)

- **The agent does the extraction itself** by running the CLI — no GEMINI or any other model, no API key required.
- **Code and markdown structure are parsed locally and deterministically** with tree-sitter AST; nothing leaves the machine. Verified in this repo: `graphify update .` built 184 nodes (markdown + shell) with no keys configured.
- The **"Tip: set GEMINI_API_KEY..."** line the CLI prints after an update is optional — **ignore it**; do not add model keys just to satisfy it.
- **Only semantic passes need a model** (deep doc understanding, images, PDFs, video — via the assistant's session model or a configured backend). For a code-first, model-free setup, skip them:
  - Key-free full rebuild: `graphify extract <path> --code-only` (local AST only, skips non-code files).
  - The CLI's own note after a code update ("For doc/paper/image changes run /graphify --update in your AI assistant") refers only to doc *semantics*; doc *structure* is already extracted locally by `graphify update .`.
- **First run / missing graph:** `graphify update .` re-extracts the full code corpus when no incremental manifest exists — it doubles as the bootstrap command.

### Semantic extraction by the agent (model-free)

The semantic pass (concepts, rationale, cross-doc relations) does **not** require GEMINI or any external model — graphify's semantic extraction is designed to be run by an AI agent (its spec literally reads "You are a graphify extraction subagent"), and the session agent can do exactly that, inline. Verified end-to-end on `docs/blueprint.md` (7 concept nodes + 7 edges + 1 hyperedge → merged graph of 191 nodes/180 edges).

1. **Read the source** — read the file(s) to understand (docs, code, papers); extract named concepts, entities, rationale, citations.
2. **Write the extraction JSON** — follow graphify's schema exactly (details + template: `docs/graphify-commands.md` → *Agent as Semantic Extractor*): `nodes` (id/label/file_type/source_file), `edges` (source/target/relation/confidence/confidence_score), optional `hyperedges`. Use `_origin: "semantic"`, `file_type` ∈ code|document|paper|image|rationale|concept, confidence EXTRACTED (1.0) / INFERRED (rubric 0.55–0.95) / AMBIGUOUS, node IDs `{stem}_{entity}` deterministic.
3. **Convert & merge** — `merge-graphs` expects graph format (nodes + **links**), so convert edges→links, then merge into the main graph:

```bash
graphify merge-graphs graphify-out/graph.json <converted>.json --out graphify-out/graph.json
```

4. **Verify** — `graphify explain "<concept>"` and `graphify path "<A>" "<B>"` against the merged graph; confirm the nodes, edges and confidence tags are queryable.

This keeps the whole pipeline offline: local AST for code + the agent for semantics — no external API keys.

### After updating

- **graphify:** outputs in `graphify-out/` — `graph.json` (queryable), `graph.html` (visual), `GRAPH_REPORT.md` (highlights). Sanity-check with `graphify god-nodes` or `graphify query "<question>"`.
- **gitnexus:** `npx gitnexus analyze` refreshes `.gitnexus/`; if the MCP tools still warn the index is stale, re-run it.

---

## Grill Mode — Planning Only, No Implementation

Active while `/grill-me` or `/grill-with-docs` is running (or the user says "grill me").

### Rules

- **Do NOT implement anything.** No code edits, no file creation for implementation, no commits of implementation work.
- Interview the user one question at a time, walking down each branch of the design tree; provide your recommended answer for each.
- If a question can be answered by exploring the codebase, explore it (graphify/gitnexus first, cat/grep/ls fallback) instead of asking.
- With `/grill-with-docs`, also sharpen domain terminology against existing docs and capture resolved decisions as they crystallise.

### Deliverables when the plan crystallises

When the interview converges, convert the agreed plan into **two artifacts**:

**1. Beads issues, each containing an implementation plan.** One issue per independently-grabbable slice (thin tracer-bullet vertical slices through all layers — schema, backend, frontend, tests). Publish blockers first so dependency IDs exist. Use `bd create` with the plan in the `--design` field:

```bash
bd create "<title>" \
  --type story \
  --design "<implementation plan for this slice>" \
  --acceptance "<acceptance criteria>" \
  --deps "blocks:<blocking-id>" \
  --labels "plan:<plan-name>" \
  --estimate <minutes>
```

**2. One MD file with the claim-order checklist**, committed at `docs/plans/<plan-name>-claim-order.md`. It lists every issue in the order agents should claim them (dependency order). Format:

```markdown
# Claim Order — <Plan Name>

Created: <date> · Source: <grill session / PRD>

Claim issues in order. Before starting, claim: `bd update <ID> --claim`.
Tick the box when claimed; mark ✅ when closed.

- [ ] bd-XX — <slice title> — blocked by: bd-YY
- [ ] bd-XX — <slice title> — can start immediately
```

Reference both artifacts in your closing summary so the user (and AFK agents) can pick up work.

### Finalization — commit & push

Grill mode produces **no code**, but its artifacts must still reach the remote so AFK agents can act on them. Commit the claim-order checklist MD together with any docs updated during the session (e.g. `CONTEXT.md`, ADRs from `/grill-with-docs`), push, and sync the beads issues:

```bash
git add docs/plans/<plan-name>-claim-order.md   # + any docs changed this session
git commit -m "plan: <plan-name> — beads issues + claim-order checklist"
git push
bd dolt push                                     # sync the newly created beads issues to remote
```

Only then is the grill session complete.

---

## Coding Mode — Implementation

Active for any task that writes or manipulates files.

### Step 0 — Activate `/ponytail` FIRST

Before writing any code, activate the **ponytail** skill (`.claude/skills/ponytail/SKILL.md`, project skill). It enforces the lazy-senior-dev ladder: YAGNI → reuse existing code → stdlib → native platform feature → already-installed dependency → one line → minimal working code. Default intensity `full`; switch with `/ponytail lite|full|ultra`. It stays active for the whole task.

Ponytail governs what you build; the Coding Rules in `docs/coding-rules.md` govern how it fits this codebase. Both apply.

### Step 0.5 — Understand before touching

- Run impact analysis on any symbol you will modify (gitnexus; fallback: grep callers).
- Trace the real flow the change touches (graphify query / explain).

### Coding Rules — Go + React (Wails)

The full rules live in **`docs/coding-rules.md`** (authoritative): file size limits, Go rules, React/TypeScript rules, verification gates. **Read it before writing code.**

#### File size limit — summary

| Language / file type | Max lines per file |
|----------------------|--------------------|
| Go (`.go`) | **300** |
| TypeScript / TSX | **300** |
| CSS | **300** |

Ceiling, not a target (~100–200 lines preferred); split on responsibility boundary, never mechanically; exemptions: generated code, migrations, fixtures, data tables, vendored code. Full detail: `docs/coding-rules.md`.

### Step last — Verify & close out

1. Run the verification gates in `docs/coding-rules.md` (go build/vet/test + frontend typecheck/build).
2. `gitnexus_detect_changes()` before committing to confirm the blast radius matches your intent.
3. Update beads: claim your issue (`bd update <ID> --claim`) before starting, close it (`bd close <ID>`) when done.
4. Follow the session-completion protocol in AGENTS.md (quality gates → close issues → push).
