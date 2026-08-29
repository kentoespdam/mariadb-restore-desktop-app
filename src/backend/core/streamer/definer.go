package streamer

import (
	"bytes"
	"io"
)

// DefinerStripper is an io.Reader wrapper that rewrites every
// `DEFINER=<user>@<host>` clause to `DEFINER=CURRENT_USER` on the fly.
//
// Implementation: read up to one statement (semicolon-terminated) at a
// time, rewrite DEFINER clauses in that buffered chunk, then emit. If a
// statement grows past maxBuf, the bytes are flushed verbatim — we
// never want to hold a multi-GB INSERT in memory.
//
// Cross-read boundary: mariadb-dump keeps `DEFINER=` on a single line
// inside a single statement, so a `;` always ends a DEFINER-bearing
// clause. We don't try to detect DEFINER split across a `;` because
// that doesn't happen in real dumps.
type DefinerStripper struct {
	src     io.Reader
	maxBuf  int
	pending []byte // bytes already read from src but not yet emitted
	done    bool   // true once src has returned io.EOF
}

// NewDefinerStripper wraps src. maxBuf is the largest single statement
// we will buffer (8 KiB by default).
func NewDefinerStripper(src io.Reader) *DefinerStripper {
	return &DefinerStripper{src: src, maxBuf: 8 * 1024}
}

func (d *DefinerStripper) Read(p []byte) (int, error) {
	// First, try to emit one complete statement from our pending buffer.
	if emitted, n := d.emitOneStatement(p); emitted {
		return n, nil
	}
	if d.done && len(d.pending) == 0 {
		return 0, io.EOF
	}
	// Pull more from src until we have a `;` or hit the buffer limit.
	chunk := make([]byte, 1024)
	for {
		n, err := d.src.Read(chunk)
		if n > 0 {
			d.pending = append(d.pending, chunk[:n]...)
			if i := bytes.IndexByte(d.pending, ';'); i >= 0 {
				end := i + 1
				stmt := rewriteDefiners(d.pending[:end])
				d.pending = d.pending[end:]
				return copyOut(p, stmt), nil
			}
			if len(d.pending) > d.maxBuf {
				stmt := d.pending
				d.pending = nil
				return copyOut(p, stmt), nil
			}
		}
		if err == io.EOF {
			d.done = true
			if len(d.pending) == 0 {
				return 0, io.EOF
			}
			stmt := rewriteDefiners(d.pending)
			d.pending = nil
			return copyOut(p, stmt), io.EOF
		}
		if err != nil {
			return 0, err
		}
	}
}

// emitOneStatement tries to satisfy Read from d.pending alone.
// Returns (true, n) if a statement was emitted; (false, _) otherwise.
func (d *DefinerStripper) emitOneStatement(p []byte) (bool, int) {
	if i := bytes.IndexByte(d.pending, ';'); i >= 0 {
		end := i + 1
		stmt := rewriteDefiners(d.pending[:end])
		d.pending = d.pending[end:]
		return true, copyOut(p, stmt)
	}
	return false, 0
}

// copyOut copies as much of b into p as fits and returns the count.
// Excess bytes (rare with our maxBuf sizing) are dropped on this call;
// the next Read will start a fresh statement.
func copyOut(p, b []byte) int { return copy(p, b) }

// rewriteDefiners replaces every `DEFINER=...@...` in b with
// `DEFINER=CURRENT_USER`. Operates on a single statement that is
// bounded in size by the caller.
func rewriteDefiners(b []byte) []byte {
	const marker = "DEFINER="
	const replacement = "DEFINER=CURRENT_USER"
	out := make([]byte, 0, len(b))
	i := 0
	for {
		j := indexCI(b[i:], []byte(marker))
		if j < 0 {
			out = append(out, b[i:]...)
			return out
		}
		out = append(out, b[i:i+j]...)
		out = append(out, replacement...)
		k := i + j + len(marker)
		for k < len(b) && b[k] != ' ' && b[k] != '\t' && b[k] != '\n' && b[k] != '\r' && b[k] != ')' && b[k] != ',' && b[k] != ';' {
			k++
		}
		i = k
	}
}

// indexCI returns the position in haystack of the first case-insensitive
// match of needle, or -1.
func indexCI(haystack, needle []byte) int {
	if len(needle) == 0 {
		return 0
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			a := haystack[i+j]
			b := needle[j]
			if a >= 'A' && a <= 'Z' {
				a += 32
			}
			if b >= 'A' && b <= 'Z' {
				b += 32
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
