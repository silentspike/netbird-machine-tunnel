//go:build !cgo

package dex

import (
	"fmt"
	"log/slog"

	"github.com/dexidp/dex/storage"
)

func openSQLiteStorage(_ *slog.Logger, _ string) (storage.Storage, error) {
	return nil, fmt.Errorf("sqlite3 storage is not available without CGO")
}
