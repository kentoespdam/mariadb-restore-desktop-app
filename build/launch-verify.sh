#!/usr/bin/env bash
# Regression test for build/launch-linux.
# Reproduces the snap-core20 GLIBC_PRIVATE symbol-lookup crash by exec'ing
# the binary (or the launcher) with LD_LIBRARY_PATH and LD_PRELOAD pointing
# at the snap core20 /lib. Without the launcher the binary dies immediately;
# with the launcher the binary starts cleanly.
#
# Usage: build/launch-verify.sh [path/to/binary] [path/to/launcher]
# Defaults:
#   binary  = build/bin/mariadb-restore-desktop-app
#   launcher= build/launch-linux

set -u

BIN=${1:-build/bin/mariadb-restore-desktop-app}
LAUNCHER=${2:-build/launch-linux}
SNAP_LIB=/snap/core20/current/lib/x86_64-linux-gnu
TIMEOUT=2

# Tiny C harness — built once, cached. Forks; the child sets the poisoned
# env then execvp's the target. The parent (and this script) keep a clean
# env, so the harness itself isn't affected by the poison.
HARNESS=/tmp/launch-verify.$$.bin
HARNESS_SRC=/tmp/launch-verify.$$.c
cat > "$HARNESS_SRC" <<'EOF'
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/wait.h>
int main(int argc, char **argv) {
    if (argc < 2) { fprintf(stderr, "usage: %s <prog>\n", argv[0]); return 2; }
    pid_t pid = fork();
    if (pid == 0) {
        setenv("LD_LIBRARY_PATH", "/snap/core20/current/lib/x86_64-linux-gnu", 1);
        setenv("LD_PRELOAD",      "/snap/core20/current/lib/x86_64-linux-gnu/libpthread.so.0", 1);
        execvp(argv[1], &argv[1]); perror("execvp"); _exit(127);
    }
    int s; waitpid(pid, &s, 0);
    if (WIFEXITED(s))   return WEXITSTATUS(s);
    if (WIFSIGNALED(s)) return 128 + WTERMSIG(s);
    return 1;
}
EOF
cc -O2 -o "$HARNESS" "$HARNESS_SRC" || { echo "FAIL: cc failed"; exit 2; }
trap 'rm -f "$HARNESS" "$HARNESS_SRC"' EXIT

if [ ! -x "$SNAP_LIB/libpthread.so.0" ]; then
  echo "SKIP: snap core20 not present at $SNAP_LIB — bug can't be reproduced here"
  exit 0
fi

echo "--- RED: $BIN with poisoned env (should fail to start) ---"
"$HARNESS" "$BIN" 2>&1 | head -2
echo ""

echo "--- GREEN: $LAUNCHER $BIN with poisoned env ---"
"$HARNESS" "$LAUNCHER" "$BIN" >/tmp/launcher-out.$$ 2>&1 &
PID=$!
sleep "$TIMEOUT"
if kill -0 "$PID" 2>/dev/null; then
  echo "PASS: binary is running (pid $PID) — launcher scrubbed the env"
  kill "$PID" 2>/dev/null
  wait "$PID" 2>/dev/null
  RC=0
else
  wait "$PID"
  echo "FAIL: launcher did not keep the binary running. Output:"
  cat "/tmp/launcher-out.$$"
  RC=1
fi
rm -f "/tmp/launcher-out.$$"
exit $RC
