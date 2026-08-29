package catalog

import "database/sql"

// sqlErrNoRows returns sql.ErrNoRows without importing database/sql
// repeatedly at call sites.
func sqlErrNoRows() error { return sql.ErrNoRows }
