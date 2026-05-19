//go:build odbc

package orm

import (
	_ "github.com/polytomic/odbc"
)

func init() {
	supportDriver[DriverAltibase] = true
}
