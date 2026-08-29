package streamer

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/scanner"
)

func TestBuildAssemblesInOrder(t *testing.T) {
	const payload = "AAAABBBBCCCCC"
	parts := []scanner.Offset{
		{StartByte: 0, EndByte: 4},  // AAAA
		{StartByte: 4, EndByte: 8},  // BBBB
		{StartByte: 8, EndByte: 13}, // CCCCC
	}
	src := strings.NewReader(payload)
	r := Build("HDR;", "FTR", src, int64(len(payload)), parts)
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	want := "HDR;AAAABBBBCCCCCFTR"
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestDefinerStripper(t *testing.T) {
	in := "CREATE TABLE t (id int) /*!50013 DEFINER=`root`@`localhost`*/;\n"
	r := NewDefinerStripper(strings.NewReader(in))
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("DEFINER=CURRENT_USER")) {
		t.Fatalf("missing replacement: %q", got)
	}
	if bytes.Contains(got, []byte("root@")) {
		t.Fatalf("definer not stripped: %q", got)
	}
}

func TestDefinerStripperCaseInsensitive(t *testing.T) {
	in := "CREATE TABLE t (id int) /*!50013 definer=`root`@`localhost`*/;\n"
	r := NewDefinerStripper(strings.NewReader(in))
	got, _ := io.ReadAll(r)
	if !bytes.Contains(got, []byte("DEFINER=CURRENT_USER")) {
		t.Fatalf("case-insensitive match missed: %q", got)
	}
}

func TestDefinerStripperPreservesContent(t *testing.T) {
	in := "INSERT INTO t VALUES (1);\nINSERT INTO t VALUES (2);\n"
	r := NewDefinerStripper(strings.NewReader(in))
	got, _ := io.ReadAll(r)
	if string(got) != in {
		t.Fatalf("rewrite changed non-definer text: %q", got)
	}
}

func TestProgressReaderCounts(t *testing.T) {
	var lastSoFar, lastTotal int64
	src := strings.NewReader("hello world")
	pr := NewProgressReader(src, 11)
	pr.OnProgress(func(soFar, total int64) {
		lastSoFar = soFar
		lastTotal = total
	})
	n, err := io.Copy(io.Discard, pr)
	if err != nil {
		t.Fatal(err)
	}
	if n != 11 {
		t.Fatalf("copied %d", n)
	}
	if pr.BytesSoFar() != 11 {
		t.Fatalf("count = %d", pr.BytesSoFar())
	}
	if lastTotal != 11 {
		t.Fatalf("lastTotal = %d", lastTotal)
	}
	if lastSoFar < 1 {
		t.Fatalf("lastSoFar = %d", lastSoFar)
	}
}
