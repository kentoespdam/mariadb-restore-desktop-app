# 0009 — Partial Restore is idempotent: synthesize DROP TABLE IF EXISTS for every selected table

Partial Restore is a **destructive replacement** operation. The selected tables on the target server are dropped and recreated from the dump before the byte ranges are streamed. This makes the operation safe to re-run against a server that already has the same tables.

## Why synthesize DROP

The mariadb-dump output this app targets does not include `DROP TABLE` statements by default (the Backup feature's "Include DROP TABLE" checkbox is unchecked by default, per ADR 0003). If a user re-runs Partial Restore against a server that already has the same tables, the `CREATE TABLE` byte range fails with `ERROR 1050 (42S01) at line N: Table 'X' already exists`. The mariadb subprocess closes stdin, every later write hits a broken pipe, and the FE sees the restore as "stuck at 0%" because no progress events arrive.

Two ways to make this idempotent were considered:

1. **Synthesize `DROP TABLE IF EXISTS` in the header** (chosen). Cheap, deterministic, no DB connectivity required at planning time. The DROP block is emitted immediately after `SET FOREIGN_KEY_CHECKS=0` so cross-table FK references don't reject the DROP. On first run, `IF EXISTS` makes the DROP a no-op.
2. **Skip `CREATE TABLE` ranges for tables that already exist on the target.** Requires the restore planner to query the target's schema, which adds a DB connection to the planning phase and a round-trip per partial restore. Rejected for the additional moving parts.

## Placement matters

The DROP block is inserted **after** `SET FOREIGN_KEY_CHECKS=0` and **before** the parts stream. The full header order is:

1. `CREATE DATABASE IF NOT EXISTS \`db\`; USE \`db\`;` (per ADR 0006)
2. `SET FOREIGN_KEY_CHECKS=0; SET UNIQUE_CHECKS=0; SET NAMES utf8mb4;`
3. `DROP TABLE IF EXISTS \`db\`.\`table\`;` for every selected `CREATE_TABLE`
4. parts (byte ranges from the catalog)
5. footer: `SET FOREIGN_KEY_CHECKS=1; COMMIT;`

## Why this only applies to Partial Restore

Full Restore passes the file through verbatim (ADR 0002). If the user wants a destructive replacement, they pick the `Include DROP TABLE` option at backup time and the resulting dump's own `DROP TABLE` statements drive the replace. Synthesizing DROP for Full Restore would change semantics for users who deliberately use a non-destructive dump.

## Consequences

- Partial Restore is now safe to re-run. The FE no longer shows "stuck 0%" on a target that already has the same tables.
- Existing tables in the selected set are wiped before the new schema is applied. Users who wanted an additive restore (only insert new rows, never touch schema) need to manage that themselves — out of scope for the v1 feature.
- The scanner must record the **table name** in `ObjectName`, not the database name. See ADR 0010.
