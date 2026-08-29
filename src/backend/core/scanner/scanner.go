package scanner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Scanner walks a dump file once and records byte ranges per object.
type Scanner struct {
	// currentDB is the database name from the most recent USE statement.
	// Empty before the first USE.
	currentDB string
}

// New returns a Scanner ready for use.
func New() *Scanner { return &Scanner{} }

// Scan opens path and returns every object it finds. Memory stays flat:
// the underlying bufio.Scanner reads line by line and the returned slice
// is the only thing held.
func (s *Scanner) Scan(path string) ([]Object, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("scanner: open: %w", err)
	}
	defer f.Close()

	// Track byte position manually. bufio.Scanner hides offsets so we
	// wrap a bufio.Reader (not Scanner) and ReadString('\n') for
	// predictable offset tracking.
	br := bufio.NewReaderSize(f, 64*1024)

	var out []Object
	var pos int64
	var pending struct {
		typ   ObjectType
		name  string
		db    string
		start int64
		open  bool // true between the opening "CREATE TABLE / INSERT" line and the closing ";"
	}

	flush := func(endPos int64) {
		if !pending.open {
			return
		}
		out = append(out, Object{
			StartByte:    pending.start,
			EndByte:      endPos,
			ObjectType:   pending.typ,
			ObjectName:   pending.name,
			DatabaseName: pending.db,
		})
		pending.open = false
	}

	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			next := pos + int64(len(line))
			trimmed := strings.TrimRight(line, "\r\n")
			upper := strings.ToUpper(trimmed)
			switch {
			case strings.HasPrefix(upper, "USE "):
				// `USE \`db\`;` or `USE db;`
				fields := strings.Fields(trimmed)
				if len(fields) >= 2 {
					name := strings.TrimSuffix(fields[1], ";")
					name = strings.Trim(name, "`")
					s.currentDB = name
				}
			case strings.HasPrefix(upper, "CREATE TABLE "):
				// start of a DDL object
				flush(pos)
				pending.typ = TypeCreateTable
				pending.name = extractIdent(trimmed, "CREATE TABLE ")
				pending.db = s.currentDB
				pending.start = pos
				pending.open = true
			case strings.HasPrefix(upper, "INSERT INTO "):
				// "INSERT INTO `t` VALUES (...);" — single-line.
				flush(pos)
				out = append(out, Object{
					StartByte:    pos,
					EndByte:      next,
					ObjectType:   TypeInsert,
					ObjectName:   extractIdent(trimmed, "INSERT INTO "),
					DatabaseName: s.currentDB,
				})
			}
			// closing semicolon for multi-line CREATE TABLE
			if pending.open && strings.HasSuffix(trimmed, ";") {
				flush(next)
			}
			pos = next
		}
		if err == io.EOF {
			flush(pos)
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("scanner: read: %w", err)
		}
	}
}

// extractIdent returns the first backtick-quoted or bare identifier that
// follows the given prefix on a line.
func extractIdent(line, prefix string) string {
	rest := strings.TrimPrefix(line, prefix)
	rest = strings.TrimLeft(rest, " \t")
	if rest == "" {
		return ""
	}
	if rest[0] == '`' {
		if i := strings.IndexByte(rest[1:], '`'); i >= 0 {
			return rest[1 : 1+i]
		}
		return ""
	}
	// bare identifier up to next space, dot, paren, or semicolon
	end := len(rest)
	for i, r := range rest {
		if r == ' ' || r == '.' || r == '(' || r == ';' || r == '\t' {
			end = i
			break
		}
	}
	return rest[:end]
}
