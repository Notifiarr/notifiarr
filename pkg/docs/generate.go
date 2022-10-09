//go:build tools
// +build tools

// Package docs provides a generator (below) and a few auto-generated files that create swag API docs.
package docs

//go:generate swag i --parseDependency --parseInternal --dir ../../ --output .

import (
	_ "github.com/swaggo/swag/cmd/swag"
)
