// Package scanner reads a mariadb-dump .sql file in a single pass and
// returns a slice of objects keyed by their byte ranges.
//
// Pure stdlib: bufio + os. No Wails, no SQLite.
package scanner

// ObjectType classifies one entry recorded by the scanner.
type ObjectType string

const (
	TypeCreateTable ObjectType = "CREATE_TABLE"
	TypeInsert      ObjectType = "INSERT"
	TypeUse         ObjectType = "USE"
)

// Offset records the byte range of one object in the dump file. Ranges are
// half-open: [StartByte, EndByte).
type Offset struct {
	StartByte    int64
	EndByte      int64
	ObjectType   ObjectType
	ObjectName   string
	DatabaseName string
}

// Object is the per-line result the scanner returns for the caller to
// persist. It is identical to Offset today; kept as its own type so the
// wire format (the catalog row) can diverge from the in-memory struct
// without breaking the public API.
type Object = Offset
