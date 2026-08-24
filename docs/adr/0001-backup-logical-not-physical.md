# 0001 — v1 backup is logical (`mariadb-dump`), not physical

MariaDB offers two backup families: logical dumps (`mariadb-dump` → `.sql` files) and physical copies (`mariadb-backup` → data files). v1 ships logical only, because the Restore pipeline (byte-offset scanner → catalog → virtual streamer) consumes `.sql` files — a logical backup closes the loop between the two workflows and stays fully portable on Windows and Linux. Physical backup would require a second, entirely different restore path (stop server, replace datadir, prepare) and is deferred.

Scheduling is out of scope for v1; backups run on demand. v1 backup granularity is whole-database: one `.sql` file per database, all objects included via `--single-transaction --routines --triggers --events`. Per-object backup selection is a v2 milestone.
