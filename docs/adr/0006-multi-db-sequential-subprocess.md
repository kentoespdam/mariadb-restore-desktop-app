# 0006 — Multi-database Partial Restore runs one subprocess per database, sequentially

When the user selects tables from more than one database in a single Partial Restore operation, the restore engine groups the selections by database and runs one mariadb CLI subprocess per database group, sequentially. Each subprocess receives its own synthetic header (`CREATE DATABASE IF NOT EXISTS … ; USE dbname;`), the selected byte ranges for that database, and a synthetic footer.

A single mixed-database stream was ruled out: it would require the Virtual Streamer to inject `USE` statements mid-stream, invalidate the `-D dbname` CLI flag, and create untraceable cross-database FK edge cases. The sequential-subprocess approach keeps the Virtual Streamer design single-database per invocation (matching the single-database Backup model) while still supporting multi-database selection in the UI.

## Consequences

Progress reporting shows `"Restoring shop_db (1/2)… Restoring inventory_db (2/2)…"`. Stop-on-first-error (ADR not yet filed) applies per subprocess: if `shop_db` fails, `inventory_db` is not attempted.
