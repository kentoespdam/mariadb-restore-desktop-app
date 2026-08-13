# Portable MariaDB Restore & Backup

The two workflows of the Portable MariaDB Restore & Backup Tool for MariaDB 10.11+: **Backup**, which produces a dump file from a live server via `mariadb-dump`, and **Restore**, which rebuilds a database from a dump file without loading it into memory — indexing it once and streaming only the user-selected parts into the `mariadb` CLI.

## Language

### Backup & restore workflows

**Full Restore**:
The Restore sub-path that pipes an entire dump file directly to the mariadb CLI without scanning; triggered when the user clicks "Restore" immediately after selecting a file, bypassing the Analyze step.
_Avoid_: direct restore, raw restore, simple restore

**Partial Restore**:
The Restore sub-path that requires an Analyze step first; the user selects individual tables and optionally their database's routines/triggers/events, then the Virtual Streamer reconstructs a stream from only the selected byte ranges.
_Avoid_: selective restore, filtered restore

**Analyze**:
The explicit user action that triggers the Byte-Offset Scanner on a selected dump file, producing a populated Catalog and enabling Partial Restore object selection.
_Avoid_: scan, index, parse

**Server Profile**:
A named set of connection credentials (host, port, user, password, SSL configuration) for one MariaDB server, stored AES-GCM-encrypted in the Catalog and selected by the user before starting Backup or Restore.
_Avoid_: connection, session, server config

**Database Section**:
The contiguous block of SQL in a multi-database dump file that belongs to one database, delimited by a `USE dbname;` statement; the Byte-Offset Scanner uses these boundaries to attribute objects to the correct database.
_Avoid_: database block, database chunk

**Backup**:
The workflow that produces a Dump File from a live MariaDB server using the `mariadb-dump` client, selecting which databases to include (all objects inside); on demand only in v1 (no scheduler).
_Avoid_: export, snapshot

**Restore**:
The workflow that rebuilds a database on a target server from a Dump File, streaming only the user-selected parts through the `mariadb` CLI.
_Avoid_: import, load

**Dump file**:
A `.sql` file produced by `mariadb-dump` containing a database's schema and data as DDL (`CREATE TABLE`) and DML (`INSERT INTO`) statements: the output of the Backup workflow and the input to the Restore workflow.
_Avoid_: SQL file, backup file, export

**Byte-Offset Scanner**:
The component that reads a dump file in a single pass without loading it into memory, recording the `start_byte`/`end_byte` coordinates of every object it contains.
_Avoid_: dump parser, dump reader, indexer

**Catalog**:
The persistent index of every object in the dump keyed by its byte range, stored in an embedded SQLite database beside the app binary.
_Avoid_: index, cache, database

**Virtual Streamer**:
The component that reconstructs a valid restore stream by concatenating the fixed header stream, the byte ranges of the user-selected objects, and the fixed footer stream, without ever materializing the selected data.
_Avoid_: merged stream, virtual file, combined reader

**Definer Stripper**:
The on-the-fly transformer that rewrites `DEFINER=` clauses to `DEFINER=CURRENT_USER` (or removes them) as the stream flows past, so a restore never requires the dump's original definer users to exist on the target server.
_Avoid_: definer replacer, definer filter

**Progress Reader**:
The wrapper around the outgoing stream that counts delivered bytes atomically and emits throttled progress events to the UI (every 100–250ms), so the UI can show a live percentage without flooding IPC.
_Avoid_: progress bar, byte counter

**mariadb-dump**:
The `mariadb-dump` command-line client subprocess that streams a live server's schema and data into a Dump File; cancelled via the process context just like the `mariadb` CLI.
_Avoid_: mysqldump, mysql-dump

**mariadb CLI**:
The `mariadb` command-line client subprocess that receives the reconstructed stream on stdin and executes the restore against the target server; cancelled via the process context so no zombie process is left behind. Located by auto-discovery (Windows Registry/PATH, `/usr/bin/mariadb` on Linux).
_Avoid_: mysql, mysql CLI, database server

### Portable storage & security

**app.key**:
The auto-generated 256-bit AES-GCM key created on first launch and stored beside the binary, used to encrypt server credentials stored in the local SQLite database.
_Avoid_: secret key, password file, encryption key

**Smart Recovery**:
The modal shown when `app.key` is missing but an existing catalog is found, offering either Cancel (so the user can recover the key) or Reset & Re-init (wipe the catalog, generate a new key, and start fresh).
_Avoid_: recovery modal, key recovery

**Executable Scope**:
The property that all application files (binary, catalog, `app.key`) live in one directory beside the executable, making the app fully portable.
_Avoid_: portable mode, app directory
