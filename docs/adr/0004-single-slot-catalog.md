# 0004 — Catalog is single-slot: every new dump file triggers a full rescan

The Catalog holds the index of exactly one dump file at a time. Opening a different dump file clears the Catalog and rescans from scratch. There is no per-file caching keyed by path, size, or mtime.

The decision prioritises correctness over convenience: a cached Catalog that drifts from the actual file on disk would produce silent data errors during restore. Target users are database administrators who understand that Analyze is a deliberate step with a cost; the alternative (multi-slot caching with invalidation logic) adds schema complexity and hidden state without a meaningful UX benefit given that Analyze is already a manual trigger.
