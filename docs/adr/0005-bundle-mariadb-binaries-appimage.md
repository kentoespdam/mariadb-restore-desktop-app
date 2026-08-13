# 0005 — mariadb and mariadb-dump are bundled; Linux ships as AppImage

`mariadb` and `mariadb-dump` are bundled inside the application distribution rather than requiring the user to install them separately. On Windows the app ships as a portable folder (zip); on Linux as an AppImage built with `linuxdeploy` + `appimagetool`.

Bundling was chosen because requiring users to install MariaDB client tools separately contradicts the Executable Scope principle (everything beside the binary, copy-and-run). A pure-Go implementation without the CLI was ruled out: reimplementing `mariadb-dump` correctly (charset, BLOB encoding, routines, DELIMITER blocks) is a separate project of its own, and statement-splitting a dump stream for native Go execution is fragile against the `DELIMITER ;;` pattern used by stored procedures and triggers.

AppImage was chosen for Linux over shipping a dynamically-linked binary (glibc version fragility across distros) or a statically-compiled binary (complex static cross-compilation of MariaDB client). AppImage bundles all native dependencies into a single self-contained executable, matching the Windows portable-folder model in spirit.

## Consequences

- GPL compliance: the About page must include a license notice and link to MariaDB source.
- CI must produce two artifacts: a Windows zip and a Linux AppImage.
- The Settings page exposes a manual override for the binary path for users who prefer their own installation.
