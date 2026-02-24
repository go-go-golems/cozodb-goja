package module

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/cozocgo"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/fakebackend"
)

func DefaultOpen(ctx context.Context, opts OpenOptions) (*cozoapi.DB, error) {
	backendName := strings.TrimSpace(opts.Backend)
	if backendName == "" {
		backendName = "fake"
	}

	policy := cozoapi.DefaultPolicy()
	policy.QueryOptions = cozoapi.QueryOptionPolicy{
		DefaultTimeoutSec: cozoapi.PtrFloat64(3),
		MaxTimeoutSec:     cozoapi.PtrFloat64(30),
		MaxLimit:          cozoapi.PtrInt64(10000),
		MaxOffset:         cozoapi.PtrInt64(100000),
	}

	switch backendName {
	case "fake":
		return cozoapi.Open(fakebackend.New(), policy)
	case "cozocgo", "cozo_cgo":
		backend, err := cozocgo.Open(ctx, cozocgo.OpenOptions{
			Engine: "mem",
		})
		if err != nil {
			return nil, err
		}
		return cozoapi.Open(backend, policy)
	default:
		return nil, fmt.Errorf("unsupported backend %q", backendName)
	}
}
