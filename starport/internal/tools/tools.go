// +build tools

package tools

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
import (
	_ "github.com/go-bindata/go-bindata"
	_ "github.com/rakyll/statik"
)
