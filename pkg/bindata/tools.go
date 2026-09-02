//go:build tools

package bindata

import (
	// Used to build windows exe metadata.
	_ "github.com/akavel/rsrc"
	// Used to create API docs (OpenAPI 3).
	_ "github.com/swaggo/swag/v2"
)
