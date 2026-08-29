// Package streamer assembles the restore stream: a fixed header, the
// user-selected byte ranges, and a fixed footer — never materializing
// the selected data in memory.
package streamer

import (
	"io"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/scanner"
)

// Build joins header, the byte ranges in parts (read from source), and
// footer into a single io.Reader. Reads are streamed; nothing is
// buffered beyond the section sizes themselves.
func Build(header, footer string, source io.ReaderAt, size int64, parts []scanner.Offset) io.Reader {
	readers := make([]io.Reader, 0, len(parts)+2)
	readers = append(readers, stringReader(header))
	for _, p := range parts {
		readers = append(readers, io.NewSectionReader(source, p.StartByte, p.EndByte-p.StartByte))
	}
	readers = append(readers, stringReader(footer))
	return io.MultiReader(readers...)
}

type sr struct {
	s   string
	off int
}

func stringReader(s string) io.Reader { return &sr{s: s} }

func (r *sr) Read(p []byte) (int, error) {
	if r.off >= len(r.s) {
		return 0, io.EOF
	}
	n := copy(p, r.s[r.off:])
	r.off += n
	return n, nil
}

var _ io.Reader = (*sr)(nil)
