//go:build tools
// +build tools

package bindata

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/kevinburke/go-bindata"
	_ "github.com/swaggo/swag/cmd/swag"
)
