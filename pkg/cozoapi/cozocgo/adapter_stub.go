//go:build !cozo_cgo

package cozocgo

import (
	"context"
	"fmt"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

type OpenOptions struct {
	Engine  string
	Path    string
	Options map[string]any
}

func Open(context.Context, OpenOptions) (cozoapi.Backend, error) {
	return nil, fmt.Errorf("cozocgo backend requires build tag cozo_cgo")
}
