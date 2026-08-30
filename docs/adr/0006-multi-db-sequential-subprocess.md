# 0006 — Multi-database Partial Restore runs in one subprocess with synthesized DDL

When the user selects tables from more than one database in a single Partial Restore operation, the restore engine launches **one** `mariadb` CLI subprocess and synthesizes the necessary `CREATE DATABASE IF NOT EXISTS` and `USE` statements in the header so the target server knows where to apply each byte range.

## Mechanism

The scanner only tracks `CREATE TABLE` and `INSERT` byte ranges — `CREATE DATABASE` and `USE` are not in the catalog. Without synthesized DDL, the first byte range that references a database that doesn't exist on the target fails with `ERROR 1049 (42000): Unknown database 'X'`, the subprocess closes stdin, and every subsequent write hits a broken pipe.

`runPartial` therefore builds a header in three layers:

1. `CREATE DATABASE IF NOT EXISTS \`db\`; USE \`db\`;` for every distinct database the parts reference (alphabetical first-seen order).
2. The standard `SET FOREIGN_KEY_CHECKS=0; SET UNIQUE_CHECKS=0; SET NAMES utf8mb4;` block (so the upcoming DROP block can ignore cross-table FKs).
3. `DROP TABLE IF EXISTS \`db\`.\`table\`;` for every selected `CREATE_TABLE` (idempotency — see ADR 0009).

The footer stays the same: `SET FOREIGN_KEY_CHECKS=1; COMMIT;`.

## Why one subprocess, not several

The original ADR (superseded) planned to group parts by database and run one subprocess per group sequentially. That design was rejected for the current implementation because:

- A single subprocess avoids the `mariadb` reconnect cost per group.
- Progress reporting remains a single `soFar`/`total` stream that maps cleanly to one progress bar.
- Cancellation via `cancel(jobID)` reaches one `exec.Cmd` instead of needing a fan-out supervisor.
- The Virtual Streamer stays single-database-per-invocation in spirit: each `parts` element still belongs to exactly one database; the synthetic DDL just ensures that database is in scope at the time its parts stream.

The original concerns (cross-database FKs, mid-stream `USE` injection, `-D dbname` flag invalidation) are addressed by `SET FOREIGN_KEY_CHECKS=0` in the header and the per-database `USE` statements in the synthesized DDL.

## Consequences

- Progress is reported once for the whole restore; users see a single bar that reflects every database being processed.
- `restore:done` with `status=success` means every part was applied; `status=error` carries the first mariadb error via `trimMsg` (see ADR 0010).
- Re-running Partial Restore against a server that already has the same tables succeeds (idempotent, per ADR 0009).
