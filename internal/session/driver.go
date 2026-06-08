// Package session provides SQLite-backed session persistence.
// This file registers the modernc.org/sqlite driver with database/sql.
package session

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Ensure the driver is registered.
var _ = sql.Drivers