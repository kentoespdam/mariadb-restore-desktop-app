# 0010 — Scanner extracts the TABLE name from `db.table` dumps; trimMsg surfaces the real mariadb error

Two related fixes landed together because they both fix how the restore planner identifies and reports problems.

## Scanner: extract the table name, not the database name

The scanner's `extractIdent` function reads the first identifier after `CREATE TABLE ` or `INSERT INTO `. mariadb-dump output uses the form `` `db`.`table` `` — the database is the first identifier, the table is the second. The original implementation returned the first identifier, so the catalog recorded `database_name = "ecommerce"` and `object_name = "ecommerce"` for every table. The UI showed every table as having the database name as its own name, and the synthesized `DROP TABLE IF EXISTS` (ADR 0009) targeted `ecommerce.ecommerce` instead of `ecommerce.users`.

`extractIdent` now walks backtick pairs from the left and returns the **last** pair's contents. For `` `db`.`table` `` it returns `table`; for bare `table` (no backticks, mysqldump `--skip-add-drop-table` style) it returns the part after the last `.` if present, otherwise the bare identifier.

The fix is covered by `TestScanBacktickDbTable` in `src/backend/core/scanner/scanner_test.go`. **Existing catalog rows recorded before this fix are stale** — the user must re-run Analyze on the dump to refresh them.

## trimMsg: surface the real mariadb error, not the `--------------` separator

The `trimMsg` function picks the first line of the mariadb stderr as the user-facing error. MariaDB echoes the offending statement followed by a row of dashes and then the `ERROR ...` line, so the first line was usually either the echoed statement or the dashes themselves. The FE then displayed messages like `--------------` which gave the user no information about what went wrong.

`trimMsg` now scans for the first occurrence of `"ERROR "` and returns the substring from there to the next newline. If stderr is silent it falls back to the subprocess exit error. The first 240 characters are kept, with a `...` suffix if longer.

## Consequences

- Catalog queries that group by `(database_name, object_name)` now produce one row per actual table, not per database.
- Re-Analyze must be triggered by the user after the fix is deployed; there is no automatic migration of the catalog rows.
- Error messages in the FE now read `ERROR 1451 (23000): Cannot delete or update a parent row...` instead of a useless separator line.
