# Graphify Command Reference

**Purpose:** primary tool for understanding this project — in **grill mode** (exploring the codebase to answer questions) and **coding mode** (tracing flows before editing). Prefer these commands over `ls`/`grep`/`cat`; fall back to shell only when the graph can't answer (see `knowledge.md` → Understanding the Code).

- Default graph: `graphify-out/graph.json` (relative to project root).
- If the graph is missing or stale: `graphify update .` first.
- Verified against graphify **v0.9.41**.

## Quick recipes

| Task | Command |
|------|---------|
| Understand architecture / hubs | `graphify god-nodes` |
| "How does X work?" | `graphify query "<question>"` |
| Explain one symbol + neighbors | `graphify explain "<Symbol>"` |
| Trace connection between two things | `graphify path "<A>" "<B>"` |
| Impact if X changes | `graphify affected "<X>"` |
| Update graph after code changes | `graphify update .` |
| Key-free full rebuild | `graphify extract . --code-only` |
| Browse graph visually | open `graphify-out/graph.html` |

## Build & Extract

| Command | Function | Model needed? |
|---------|----------|---------------|
| `graphify update <path>` | Incremental re-extraction of **code only**; `--force` (overwrite even if nodes decreased), `--no-cluster` | **No** (local AST) |
| `graphify extract <path>` | Headless full pipeline (AST + semantic) for CI/scripts. Flags: `--backend` (gemini\|kimi\|claude\|openai\|deepseek\|ollama), `--model`, `--mode deep`, `--force`, `--max-workers`, `--token-budget`, `--max-concurrency`, `--api-timeout`, `--out`, `--no-gitignore`, `--no-cluster`, `--code-only`, `--postgres DSN`, `--cargo`, `--global`, `--as` | Only for semantic files unless `--code-only` |
| `graphify watch <path>` | Watch folder; auto-rebuild on code changes (no LLM); doc/media changes write a `needs_update` flag | Code: no. Docs/media: yes |
| `graphify cluster-only <path>` | Rerun clustering (Leiden) + report. Flags: `--resolution`, `--exclude-hubs`, `--min-community-size`, `--no-label`, `--backend`, `--model`, `--no-viz` | Clustering: no. Community naming: yes |
| `graphify label <path>` | (Re)name communities with an LLM backend. Flags: `--missing-only`, `--backend`, `--model`, `--max-concurrency`, `--batch-size` | Yes (labeling) |
| `graphify check-update <path>` | Cron-safe check of the `needs_update` flag | No |

## Query & Navigate

| Command | Function | Flags |
|---------|----------|-------|
| `graphify query "<question>"` | BFS traversal answering a plain-language question | `--dfs`, `--context C` (edge-context filter, repeatable), `--budget N`, `--graph` |
| `graphify path "A" "B"` | Shortest path between two nodes | `--graph` |
| `graphify explain "X"` | Plain-language explanation of a node + its neighbors (edges tagged EXTRACTED / INFERRED / AMBIGUOUS) | `--graph` |
| `graphify affected "X"` | Reverse traversal: what breaks if X changes | `--relation R`, `--depth N`, `--graph` |
| `graphify god-nodes` | Most-connected nodes (architectural hubs) | `--top N`, `--json`, `--graph` |
| `graphify tree` | D3 collapsible-tree HTML (`graphify-out/GRAPH_TREE.html`) | `--graph`, `--output`, `--root`, `--max-children`, `--top-k-edges`, `--label` |

## Export & Analysis

| Command | Function |
|---------|----------|
| `graphify export <format>` | Formats: `html`, `callflow-html` (Mermaid architecture/call-flow), `obsidian`, `wiki`, `svg`, `graphml`, `neo4j`, `falkordb`. Neo4j/FalkorDB support `--push URI --user U --password P` (or env `NEO4J_PASSWORD` / `FALKORDB_PASSWORD`) |
| `graphify benchmark [graph.json]` | Measure token reduction vs naive full-corpus approach |
| `graphify diagnose multigraph` | Same-endpoint edge-collapse risk report; `--json`, `--max-examples`, `--directed`/`--undirected`, `--extract-path` |

## Ingest & Multi-repo

| Command | Function |
|---------|----------|
| `graphify add <url>` | Fetch URL → `./raw/` → update graph; `--author`, `--contributor`, `--dir` |
| `graphify clone <github-url>` | Clone repo locally and print its path for `/graphify` |
| `graphify merge-graphs <g1> <g2>` | Merge two+ graphs into one cross-repo graph; `--out`, `--branch` |
| `graphify merge-driver` | Git merge driver (union-merge graph.json); set up via `graphify hook install` |
| `graphify global add\|remove\|list\|path` | Global graph at `~/.graphify/global-graph.json`; `--as <tag>` |

## Memory / Feedback Loop

| Command | Function |
|---------|----------|
| `graphify save-result` | Save a Q&A result to `graphify-out/memory/`; `--question`, `--answer`, `--type` (query\|path_query\|explain), `--nodes`, `--outcome` (useful\|dead_end\|corrected), `--correction`, `--memory-dir` |
| `graphify reflect` | Aggregate memory outcomes into a deterministic lessons doc (`graphify-out/reflections/LESSONS.md`); `--memory-dir`, `--out`, `--graph`, `--half-life-days`, `--min-corroboration` |

## Hooks & Install

| Command | Function |
|---------|----------|
| `graphify hook install\|uninstall\|status` | post-commit/post-checkout auto-rebuild (AST only, no API cost) + merge driver |
| `graphify install [--platform P]` | Register the `/graphify` skill with your assistant. Platforms: `claude\|windows\|codebuddy\|codex\|opencode\|aider\|amp\|agents\|claw\|droid\|trae\|trae-cn\|gemini\|cursor\|antigravity\|hermes\|kiro\|pi\|devin`. Flags: `--project` (per-repo install), `--strict` (Claude Code: force graph-first) |
| `graphify uninstall [--purge]` | Remove graphify from all detected platforms; `--purge` also deletes `graphify-out/` |
| `<platform> install\|uninstall` | Per-platform variants: `gemini`, `cursor`, `claude`, `codebuddy`, `codex`, `opencode`, `kilo`, `aider`, `copilot`, `vscode`, `claw`, `droid`, `trae`, `trae-cn`, `antigravity`, `hermes`, `kiro`, `pi`, `devin` |

## Hidden / Skill-level

- **`graphify prs`** — PR dashboard (exists in v0.9.41 but not listed in `--help`): `prs` (dashboard), `prs <number>` (deep dive), `--triage` (AI-ranked review queue), `--worktrees`, `--conflicts` (PRs sharing graph communities), `--base <branch>`.
- **Skill-level flags** (run by the assistant via `/graphify`, not all exposed as CLI): `--mode deep`, `--directed`, `--whisper-model`, `--no-viz`, `--svg`, `--graphml`, `--neo4j[--push]`, `--falkordb[--push]`, `--mcp` (start MCP stdio server), `--watch`, `--wiki`, `--obsidian --obsidian-dir`.

## MCP Server (agent access)

```bash
python -m graphify.serve graphify-out/graph.json                          # stdio (default)
python -m graphify.serve graphify-out/graph.json --transport http --port 8080   # shared HTTP
```

## Agent as Semantic Extractor (no external LLM)

The semantic pass can be performed by the agent itself — no GEMINI/API keys. This mirrors how graphify's own skill works in IDE mode (the assistant's model is the extractor); the session agent does the same thing inline, fully offline.

### 1. Read

Read the file(s) to understand (docs, code, papers). Extract **named concepts and entities**, rationale (WHY decisions, trade-offs), citations, and relations between them. Don't re-extract imports — AST already has those.

### 2. Write the extraction JSON (graphify schema)

```json
{
  "nodes": [
    {"id": "docs_blueprint_fastbyteoffsetscanner", "label": "Fast Byte-Offset Scanner", "file_type": "concept", "source_file": "docs/blueprint.md", "source_location": null}
  ],
  "edges": [
    {"source": "docs_blueprint_fastbyteoffsetscanner", "target": "docs_blueprint_portablecatalog", "relation": "shares_data_with", "confidence": "EXTRACTED", "confidence_score": 1.0, "source_file": "docs/blueprint.md", "weight": 1.0}
  ],
  "hyperedges": [
    {"id": "docs_blueprint_restorepipeline", "label": "Restore Pipeline", "nodes": ["n1", "n2", "n3"], "relation": "form", "confidence": "EXTRACTED", "confidence_score": 1.0, "source_file": "docs/blueprint.md"}
  ],
  "input_tokens": 0,
  "output_tokens": 0
}
```

**Schema rules:**
- `file_type` MUST be exactly one of: `code`, `document`, `paper`, `image`, `rationale`, `concept`.
- `relation` ∈ `calls | implements | references | cites | conceptually_related_to | shares_data_with | semantically_similar_to | rationale_for`.
- `confidence_score` is REQUIRED on every edge: EXTRACTED = 1.0; INFERRED = pick ONE of {0.95, 0.85, 0.75, 0.65, 0.55} (never 0.5); AMBIGUOUS = 0.1–0.3.
- **Node ID:** lowercase, only `[a-z0-9_]`. Format `{stem}_{entity}` — stem = repo-relative path without extension, every segment joined with `_` (e.g. `docs/blueprint.md` + "Fast Byte-Offset Scanner" → `docs_blueprint_fastbyteoffsetscanner`). IDs must be deterministic from the label alone; never append chunk numbers.
- `source_file`: the originating path verbatim.
- `calls` edges: source = caller, target = callee; must stay within one language (no cross-language `calls`).
- Rationale (WHY): store as a `rationale` attribute on the relevant concept node — do NOT create separate rationale nodes.
- Hyperedges: max 3 per chunk, `relation` ∈ `participate_in | implement | form`.

### 3. Convert to graph format & merge

`graphify merge-graphs` expects **graph format** (`nodes` + `links`), not the extraction format (`nodes` + `edges`). Convert first, tagging provenance:

```python
import json
ext = json.load(open('graphify-out/agent-semantic-extract.json', encoding='utf-8'))
g = {'directed': False, 'multigraph': False, 'graph': {}, 'nodes': [], 'links': [], 'hyperedges': ext.get('hyperedges', [])}
for n in ext['nodes']:
    n['_origin'] = 'semantic'; n['norm_label'] = n['label'].lower(); g['nodes'].append(n)
for e in ext['edges']:
    e['_origin'] = 'semantic'; e['context'] = e['relation']; g['links'].append(e)
json.dump(g, open('graphify-out/agent-semantic-graph.json', 'w', encoding='utf-8'), ensure_ascii=False)
```

```bash
graphify merge-graphs graphify-out/graph.json graphify-out/agent-semantic-graph.json --out graphify-out/graph.json
```

### 4. Verify

```bash
graphify explain "<concept>" --graph graphify-out/graph.json
graphify path "<A>" "<B>" --graph graphify-out/graph.json
graphify query "<question>" --graph graphify-out/graph.json
```

**Verified** (v0.9.41): `docs/blueprint.md` → 7 concept nodes + 7 edges (EXTRACTED/INFERRED) + 1 hyperedge; merged into a 191-node / 180-edge graph; `explain` and `path` both confirmed the pipeline (Fast Byte-Offset Scanner → Portable Catalog → Virtual Streamer → mariadb CLI) with confidence tags.

## Model dependency — what needs an LLM

- **Never needs a model:** `update`, `watch` (code), `extract --code-only`, `query`, `path`, `explain`, `affected`, `god-nodes`, clustering itself. Code and markdown **structure** are parsed locally & deterministically with tree-sitter AST — nothing leaves the machine, no API keys.
- **Needs a model (semantic pass):** deep doc understanding, images, PDFs, video/audio, community *naming* (`label`, or `cluster-only --backend`). In an IDE session the assistant's own model is used; headless it needs a backend key (`GEMINI_API_KEY`, `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, Ollama, etc.).
- **For this project (code-first, model-free):** stick to `graphify update .` + `extract --code-only`. Ignore the CLI's "Tip: set GEMINI_API_KEY..." hint.
