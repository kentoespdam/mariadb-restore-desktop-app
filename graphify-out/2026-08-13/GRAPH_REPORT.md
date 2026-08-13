# Graph Report - mariadb-restore-desktop-app  (2026-08-13)

## Corpus Check
- 7 files · ~6,525 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 109 nodes · 98 edges · 15 communities (10 shown, 5 thin omitted)
- Extraction: 87% EXTRACTED · 13% INFERRED · 0% AMBIGUOUS · INFERRED: 13 edges (avg confidence: 0.77)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `42f61532`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Project Instructions for AI Agents
- Agent Instructions
- Knowledge — Agent Operating Manual
- Virtual Streamer
- Graphify
- Ponytail Ladder
- 🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)
- Coding Mode — Implementation
- Coding Rules — Go + React (Wails)
- Context-First Go
- Local-First State
- Effect Cleanup
- Go Error Wrapping
- 🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)
- Table-Driven Go Tests

## God Nodes (most connected - your core abstractions)
1. `Graphify` - 8 edges
2. `Project Instructions for AI Agents` - 7 edges
3. `Agent Instructions` - 6 edges
4. `Knowledge — Agent Operating Manual` - 6 edges
5. `Coding Rules — Go + React (Wails)` - 6 edges
6. `GitNexus — Code Intelligence` - 5 edges
7. `GitNexus — Code Intelligence` - 5 edges
8. `🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)` - 5 edges
9. `🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)` - 5 edges
10. `Coding Mode — Implementation` - 5 edges

## Surprising Connections (you probably didn't know these)
- `Graphify Affected` --references--> `Graphify`  [INFERRED]
  docs/graphify-commands.md → knowledge.md
- `Graphify Explain` --references--> `Graphify`  [INFERRED]
  docs/graphify-commands.md → knowledge.md
- `Graphify MCP Server` --references--> `Graphify`  [INFERRED]
  docs/graphify-commands.md → knowledge.md
- `Graphify Path` --references--> `Graphify`  [INFERRED]
  docs/graphify-commands.md → knowledge.md
- `Graphify Query` --references--> `Graphify`  [EXTRACTED]
  docs/graphify-commands.md → knowledge.md

## Hyperedges (group relationships)
- **Restore Pipeline** — repo-2::docs_blueprint_fastbyteoffsetscanner, repo-2::docs_blueprint_portablecatalog, repo-2::docs_blueprint_virtualstreamer, repo-2::docs_blueprint_progressreader, repo-2::docs_blueprint_mariadbcli [EXTRACTED 1.00]
- **Agent Workflow Modes** — repo-2::knowledge_operatingmanual, repo-2::knowledge_grillmode, repo-2::knowledge_codingmode, repo-2::knowledge_graphupdateprocess, repo-2::knowledge_agentsemanticextraction [EXTRACTED 1.00]

## Communities (15 total, 5 thin omitted)

### Community 0 - "Project Instructions for AI Agents"
Cohesion: 0.13
Nodes (14): Always Do, Architecture Overview, Beads Issue Tracker, Build & Test, CLI, Conventions & Patterns, GitNexus — Code Intelligence, Never Do (+6 more)

### Community 1 - "Agent Instructions"
Cohesion: 0.14
Nodes (13): Agent Instructions, Always Do, Beads Issue Tracker, CLI, GitNexus — Code Intelligence, Never Do, Non-Interactive Shell Commands, Operating Manual — READ FIRST (+5 more)

### Community 2 - "Knowledge — Agent Operating Manual"
Cohesion: 0.17
Nodes (11): 1. graphify — codebase graph, 2. gitnexus — call-graph intelligence, 3. Fallback — cat / grep / ls, Deliverables when the plan crystallises, Finalization — commit & push, Grill Mode — Planning Only, No Implementation, Knowledge — Agent Operating Manual, Modes (+3 more)

### Community 3 - "Virtual Streamer"
Cohesion: 0.18
Nodes (12): app.key (AES-256-GCM), Definer Stripper, Fast Byte-Offset Scanner, mariadb CLI Subprocess, Portable Catalog, ProgressReader, Session Flags Header, Smart Recovery Modal (+4 more)

### Community 4 - "Graphify"
Cohesion: 0.20
Nodes (11): Graphify Affected, Graphify Explain, Graphify MCP Server, Merge-Graphs, Graphify Path, Graphify Query, Agent Semantic Extraction, cat/grep/ls Fallback (+3 more)

### Community 5 - "Ponytail Ladder"
Cohesion: 0.25
Nodes (9): Claim-Order Checklist Loop, 300-Line File Cap, Ponytail Ladder, Go Stdlib-First, Verification Gates, Claim-Order Checklist, Coding Mode, Grill Mode (+1 more)

### Community 6 - "🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)"
Cohesion: 0.33
Nodes (5): 1. Storage & Portable Security Layer, 2. Dump Analysis & Byte-Offset Indexer, 3. Execution Engine & On-The-Fly Stream Filtering, 4. Progress Tracking & Graceful Cancellation, 🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)

### Community 7 - "Coding Mode — Implementation"
Cohesion: 0.33
Nodes (6): Coding Mode — Implementation, Coding Rules — Go + React (Wails), File size limit — summary, Step 0.5 — Understand before touching, Step 0 — Activate `/ponytail` FIRST, Step last — Verify & close out

### Community 8 - "Coding Rules — Go + React (Wails)"
Cohesion: 0.18
Nodes (10): 1. Before starting — claim it, 2. Do the work, 3. After finishing — close issue + update checklist, 4. Finalization — commit & push, Coding Rules — Go + React (Wails), File Size Limit (hard ceiling), Go, React / TypeScript (Wails frontend) (+2 more)

### Community 13 - "🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)"
Cohesion: 0.33
Nodes (5): 1. Storage & Portable Security Layer, 2. Dump Analysis & Byte-Offset Indexer, 3. Execution Engine & On-The-Fly Stream Filtering, 4. Progress Tracking & Graceful Cancellation, 🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)

## Knowledge Gaps
- **70 isolated node(s):** `Operating Manual — READ FIRST`, `Quick Reference`, `Non-Interactive Shell Commands`, `Quick Reference`, `Rules` (+65 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Knowledge — Agent Operating Manual` connect `Knowledge — Agent Operating Manual` to `Coding Mode — Implementation`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Coding Mode — Implementation` connect `Coding Mode — Implementation` to `Knowledge — Agent Operating Manual`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Are the 5 inferred relationships involving `Graphify` (e.g. with `Graphify Affected` and `Graphify Explain`) actually correct?**
  _`Graphify` has 5 INFERRED edges - model-reasoned connections that need verification._
- **What connects `Operating Manual — READ FIRST`, `Quick Reference`, `Non-Interactive Shell Commands` to the rest of the system?**
  _70 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Project Instructions for AI Agents` be split into smaller, more focused modules?**
  _Cohesion score 0.13333333333333333 - nodes in this community are weakly interconnected._
- **Should `Agent Instructions` be split into smaller, more focused modules?**
  _Cohesion score 0.14285714285714285 - nodes in this community are weakly interconnected._