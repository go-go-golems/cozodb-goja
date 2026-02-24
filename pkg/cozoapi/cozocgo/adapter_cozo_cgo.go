//go:build cozo_cgo

package cozocgo

import (
	"context"
	"fmt"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

// TODO(manuel): wire this adapter to an actual cozo-lib-go binding package.
func Open(context.Context, OpenOptions) (cozoapi.Backend, error) {
	return nil, fmt.Errorf("cozocgo backend not implemented yet")
}
