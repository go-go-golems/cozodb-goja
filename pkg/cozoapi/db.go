package cozoapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var ErrClosed = errors.New("cozoapi db is closed")

type Backend interface {
	Name() string
	Capabilities() BackendCapabilities
	Exec(ctx context.Context, script string, params CozoParams, opts QueryOptions) (CozoResult, error)
	Export(ctx context.Context, relations []string) (map[string]RelationRows, error)
	Import(ctx context.Context, data map[string]RelationRows) error
	Close(ctx context.Context) error
}

type DB struct {
	backend Backend
	policy  Policy

	mu     sync.RWMutex
	closed bool
}

func Open(backend Backend, policy Policy) (*DB, error) {
	if backend == nil {
		return nil, fmt.Errorf("backend is required")
	}
	if policy.DeniedRelation == nil {
		policy.DeniedRelation = map[string]struct{}{}
	}
	return &DB{
		backend: backend,
		policy:  policy,
	}, nil
}

func (db *DB) Backend() string {
	if db == nil || db.backend == nil {
		return ""
	}
	return db.backend.Name()
}

func (db *DB) Capabilities() BackendCapabilities {
	if db == nil || db.backend == nil {
		return BackendCapabilities{}
	}
	return db.backend.Capabilities()
}

func (db *DB) ExecScript(ctx context.Context, script string, params CozoParams, opts *QueryOptions) (CozoResult, error) {
	query := PreparedQuery{
		Script: script,
		Params: params,
		Opts:   opts,
		Meta: QueryMeta{
			UsesSystemOp: strings.Contains(script, "::"),
		},
	}
	return db.ExecPrepared(ctx, query)
}

func (db *DB) ExecPrepared(ctx context.Context, query PreparedQuery) (CozoResult, error) {
	if err := db.ensureOpen(); err != nil {
		return CozoResult{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(query.Script) == "" {
		return CozoResult{}, fmt.Errorf("script cannot be empty")
	}
	if query.Params == nil {
		query.Params = CozoParams{}
	}

	enforced, err := db.policy.Enforce(query)
	if err != nil {
		return CozoResult{}, err
	}

	opts := QueryOptions{}
	if enforced.Opts != nil {
		opts = enforced.Opts.Clone()
	}
	return db.backend.Exec(ctx, strings.TrimSpace(enforced.Script), enforced.Params, opts)
}

func (db *DB) Q(ctx context.Context, parts []string, values []CozoValue, opts *QueryOptions) (CozoResult, error) {
	query, err := CompileTemplate(parts, values, opts)
	if err != nil {
		return CozoResult{}, err
	}
	return db.ExecPrepared(ctx, query)
}

func (db *DB) CQ(parts []string, values []CozoValue, opts *QueryOptions) (PreparedQuery, error) {
	return CompileTemplate(parts, values, opts)
}

func (db *DB) Atomic(ctx context.Context, queries []PreparedQuery, opts *AtomicOptions) (CozoResult, error) {
	query, err := CompileAtomic(queries, opts)
	if err != nil {
		return CozoResult{}, err
	}
	return db.ExecPrepared(ctx, query)
}

func (db *DB) Rel(name string) *RelationHandle {
	return &RelationHandle{db: db, name: name}
}

func (db *DB) Export(ctx context.Context, relations []string) (map[string]RelationRows, error) {
	if err := db.ensureOpen(); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return db.backend.Export(ctx, relations)
}

func (db *DB) Import(ctx context.Context, data map[string]RelationRows) error {
	if err := db.ensureOpen(); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return db.backend.Import(ctx, data)
}

func (db *DB) Close(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return nil
	}
	db.closed = true
	if ctx == nil {
		ctx = context.Background()
	}
	return db.backend.Close(ctx)
}

func (db *DB) ensureOpen() error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.closed {
		return ErrClosed
	}
	return nil
}
