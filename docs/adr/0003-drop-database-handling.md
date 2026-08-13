# 0003 — DROP DATABASE is stripped in Partial Restore, passed through in Full Restore

`DROP TABLE IF EXISTS` is always passed through (per-table replacement is expected behaviour for both paths). `DROP DATABASE IF EXISTS` is treated differently by path: in Partial Restore it is always stripped by the Stream Transformer because the user is restoring to an existing database and the nuclear option must never execute silently; in Full Restore it is passed through because the user has explicitly chosen to restore the entire file and presumably intends full replacement.

The Backup workflow exposes a checkbox "Include DROP DATABASE" (`--add-drop-database`) that defaults to unchecked, so dumps produced by this app are safe by default even in Full Restore.

## Considered Options

Stripping in both paths was rejected because it would make Full Restore unable to cleanly replace an existing database. Passing through in both paths was rejected because Partial Restore users are restoring to an existing database where a stray `DROP DATABASE` would destroy data they did not intend to lose.
