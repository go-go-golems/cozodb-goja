package fakebackend

import (
	"context"
	"sync"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

type ExecCall struct {
	Script string
	Params cozoapi.CozoParams
	Opts   cozoapi.QueryOptions
}

type queuedResult struct {
	Result cozoapi.CozoResult
	Err    error
}

type Backend struct {
	NameValue         string
	CapabilitiesValue cozoapi.BackendCapabilities

	mu      sync.Mutex
	calls   []ExecCall
	results []queuedResult
	data    map[string]cozoapi.RelationRows
	closed  bool
}

var _ cozoapi.Backend = (*Backend)(nil)

func New() *Backend {
	return &Backend{
		NameValue: "fake",
		data:      map[string]cozoapi.RelationRows{},
	}
}

func (b *Backend) Name() string {
	if b.NameValue != "" {
		return b.NameValue
	}
	return "fake"
}

func (b *Backend) Capabilities() cozoapi.BackendCapabilities {
	return b.CapabilitiesValue
}

func (b *Backend) PushExecResult(result cozoapi.CozoResult, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.results = append(b.results, queuedResult{Result: result, Err: err})
}

func (b *Backend) Calls() []ExecCall {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]ExecCall, len(b.calls))
	copy(out, b.calls)
	return out
}

func (b *Backend) Exec(_ context.Context, script string, params cozoapi.CozoParams, opts cozoapi.QueryOptions) (cozoapi.CozoResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	clonedParams := cozoapi.CozoParams{}
	for k, v := range params {
		clonedParams[k] = v
	}
	b.calls = append(b.calls, ExecCall{
		Script: script,
		Params: clonedParams,
		Opts:   opts.Clone(),
	})

	if len(b.results) == 0 {
		return cozoapi.CozoResult{}, nil
	}
	next := b.results[0]
	b.results = b.results[1:]
	return next.Result, next.Err
}

func (b *Backend) Export(_ context.Context, relations []string) (map[string]cozoapi.RelationRows, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := map[string]cozoapi.RelationRows{}
	for _, relation := range relations {
		if rows, ok := b.data[relation]; ok {
			out[relation] = rows
		}
	}
	return out, nil
}

func (b *Backend) Import(_ context.Context, data map[string]cozoapi.RelationRows) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for relation, rows := range data {
		b.data[relation] = rows
	}
	return nil
}

func (b *Backend) Close(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	return nil
}

func (b *Backend) Closed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.closed
}
