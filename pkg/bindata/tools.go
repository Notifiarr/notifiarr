//go:build tools
// +build tools

package bindata

import (
	// Used to build windows exe metadata.
	_ "github.com/akavel/rsrc"
	// Used to convert static files to internal binary data.
	_ "github.com/kevinburke/go-bindata"
	// Used to create API docs.
	_ "github.com/swaggo/swag"
)
