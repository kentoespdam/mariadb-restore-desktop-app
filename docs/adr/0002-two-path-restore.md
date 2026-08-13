# 0002 — Restore has two distinct paths: Full and Partial

Restore splits into two paths at the UI level: Full Restore (user selects file → clicks "Restore" immediately, the entire dump is piped to the mariadb CLI without scanning) and Partial Restore (user clicks "Analyze" first, the Byte-Offset Scanner runs, then the user selects objects before clicking "Restore"). The split exists because scan cost on a 30 GB file can be minutes — paying that cost when the user just wants the whole database is wasteful friction. Full Restore follows MySQL Administrator's behavior; Partial Restore is the differentiating feature.

## Consequences

The Catalog and Virtual Streamer are only exercised in the Partial path. The Full path is a simpler pipeline: file stream → Stream Transformers (Definer Stripper only) → Progress Reader → mariadb CLI subprocess.
