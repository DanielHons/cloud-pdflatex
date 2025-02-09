//go:build (darwin && cgo) || linux
// +build darwin,cgo linux

package main

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
