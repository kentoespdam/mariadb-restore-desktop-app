# Issue tracker: Beads (bd)

Issues and specs for this repo live in **beads** (`bd`), a git-native issue tracker
stored in a local Dolt database (`.beads/`). Use the `bd` CLI for all operations —
never `gh`, never markdown files. Run `bd prime` for the full workflow context.

## Conventions

- **Create an issue**: `bd create "<title>"` — add `--design "..."`, `--acceptance "..."`, `--deps "blocks:<id>"`, `--labels "..."`, `--estimate <minutes>` as appropriate
- **List issues**: `bd list`; find available work with `bd ready`
- **Read an issue**: `bd show <id>` (thread: `bd comments <id>`)
- **Comment / note**: `bd comment <id> "text"`; append a note with `bd note <id> "text"`
- **Claim**: `bd update <id> --claim` (atomic; sets assignee + in_progress)
- **Dependencies**: `--deps "blocks:<blocker-id>"` at create time, or `bd link <child-id> <blocker-id>` (id2 blocks id1 by default)
- **Labels**: `bd label` / `bd tag`; add at create time with `--labels`
- **Close**: `bd close <id>`
- **Sync**: `bd dolt push` / `bd dolt pull` — syncs the Dolt DB to the git remote under `refs/dolt/data`

Storage: issues live in the local Dolt DB under `.beads/`; `.beads/issues.jsonl` is a
passive export, not the source of truth, and `bd import` is not used during normal operation.

## When a skill says "publish to the issue tracker"

Run `bd create "<title>"` with `--design` / `--acceptance` / `--deps` / `--labels` / `--estimate` as appropriate.

## When a skill says "fetch the relevant ticket"

Run `bd show <id>` (with `bd comments <id>` for the full thread).
