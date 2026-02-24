package cozoapi_test

import (
	"context"
	"strings"
	"testing"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/fakebackend"
)

func TestResultHelpers(t *testing.T) {
	count := int64(2)
	result := cozoapi.CozoResult{
		Headers: []string{"id", "name"},
		Rows: []cozoapi.CozoRow{
			{"u1", "Ada"},
			{"u2", "Lin"},
		},
		Count: &count,
	}

	objects := result.Objects()
	if len(objects) != 2 {
		t.Fatalf("objects length = %d, want 2", len(objects))
	}
	if objects[0]["id"] != "u1" || objects[0]["name"] != "Ada" {
		t.Fatalf("objects[0] = %#v", objects[0])
	}

	first := result.FirstObject()
	if first["id"] != "u1" || first["name"] != "Ada" {
		t.Fatalf("first object = %#v", first)
	}

	if got := result.Scalar(); got != "u1" {
		t.Fatalf("scalar = %#v, want %#v", got, "u1")
	}
}

func TestNormalizeQueryOptions(t *testing.T) {
	opts := cozoapi.QueryOptions{
		Limit:      cozoapi.PtrInt64(-8),
		Offset:     cozoapi.PtrInt64(100),
		TimeoutSec: cozoapi.PtrFloat64(45),
	}
	policy := cozoapi.QueryOptionPolicy{
		DefaultTimeoutSec: cozoapi.PtrFloat64(3),
		MaxTimeoutSec:     cozoapi.PtrFloat64(10),
		MaxLimit:          cozoapi.PtrInt64(50),
		MaxOffset:         cozoapi.PtrInt64(20),
	}

	got := cozoapi.NormalizeQueryOptions(&opts, policy)
	if got.Limit == nil || *got.Limit != 0 {
		t.Fatalf("limit = %v", got.Limit)
	}
	if got.Offset == nil || *got.Offset != 20 {
		t.Fatalf("offset = %v", got.Offset)
	}
	if got.TimeoutSec == nil || *got.TimeoutSec != 10 {
		t.Fatalf("timeout = %v", got.TimeoutSec)
	}
}

func TestCompileTemplate(t *testing.T) {
	query, err := cozoapi.CompileTemplate(
		[]string{"?[id] := *users{id: ", "}", ""},
		[]cozoapi.CozoValue{"u1", 10},
		&cozoapi.QueryOptions{Limit: cozoapi.PtrInt64(5)},
	)
	if err != nil {
		t.Fatalf("compile template: %v", err)
	}

	if want := "?[id] := *users{id: $p1}$p2"; query.Script != want {
		t.Fatalf("script = %q, want %q", query.Script, want)
	}
	if query.Params["p1"] != "u1" || query.Params["p2"] != 10 {
		t.Fatalf("params = %#v", query.Params)
	}
	if query.Opts == nil || query.Opts.Limit == nil || *query.Opts.Limit != 5 {
		t.Fatalf("opts = %#v", query.Opts)
	}
}

func TestCompileAtomicNamespacesParams(t *testing.T) {
	query, err := cozoapi.CompileAtomic([]cozoapi.PreparedQuery{
		{
			Script: "?[id] <- [[$id]]",
			Params: cozoapi.CozoParams{"id": "u1"},
			Meta:   cozoapi.QueryMeta{Relations: []string{"users"}},
		},
		{
			Script: "?[id] <- [[$id]]",
			Params: cozoapi.CozoParams{"id": "u2"},
			Meta:   cozoapi.QueryMeta{Relations: []string{"users", "audit"}},
		},
	}, nil)
	if err != nil {
		t.Fatalf("compile atomic: %v", err)
	}
	if !strings.Contains(query.Script, "$q1_id") || !strings.Contains(query.Script, "$q2_id") {
		t.Fatalf("script = %q", query.Script)
	}
	if query.Params["q1_id"] != "u1" || query.Params["q2_id"] != "u2" {
		t.Fatalf("params = %#v", query.Params)
	}
	if len(query.Meta.Relations) != 2 {
		t.Fatalf("relations = %#v", query.Meta.Relations)
	}
}

func TestCompileRelationHelpers(t *testing.T) {
	create, err := cozoapi.CompileRelationCreate("users", cozoapi.RelationSpec{
		Keys:   map[string]string{"id": "String"},
		Values: map[string]string{"name": "String"},
	}, cozoapi.RelationCreateOptions{})
	if err != nil {
		t.Fatalf("create compile: %v", err)
	}
	if !strings.Contains(create.Script, ":create users") {
		t.Fatalf("create script = %q", create.Script)
	}

	put, err := cozoapi.CompileRelationPut("users", []map[string]cozoapi.CozoValue{
		{"id": "u1", "name": "Ada"},
		{"id": "u2", "name": "Lin"},
	}, cozoapi.RelationMutationOptions{Returning: true})
	if err != nil {
		t.Fatalf("put compile: %v", err)
	}
	if !strings.Contains(put.Script, ":put users") || !strings.Contains(put.Script, ":returning") {
		t.Fatalf("put script = %q", put.Script)
	}

	access, err := cozoapi.CompileRelationAccess("users", cozoapi.AccessReadOnly)
	if err != nil {
		t.Fatalf("access compile: %v", err)
	}
	if access.Meta.UsesSystemOp != true || !strings.Contains(access.Script, "::access_level users read_only") {
		t.Fatalf("access query = %#v", access)
	}
}

func TestPolicyEnforcement(t *testing.T) {
	policy := cozoapi.DefaultPolicy()
	policy.AllowSystemOps = false
	policy = policy.WithDeniedRelations("secrets")
	policy.QueryOptions = cozoapi.QueryOptionPolicy{
		MaxLimit:      cozoapi.PtrInt64(10),
		MaxTimeoutSec: cozoapi.PtrFloat64(2),
	}

	_, err := policy.Enforce(cozoapi.PreparedQuery{
		Script: "::columns users",
		Params: cozoapi.CozoParams{},
	})
	if err == nil {
		t.Fatalf("expected system op rejection")
	}

	_, err = policy.Enforce(cozoapi.PreparedQuery{
		Script: "?[x] := *secrets{x}",
		Meta: cozoapi.QueryMeta{
			Relations: []string{"secrets"},
		},
	})
	if err == nil {
		t.Fatalf("expected relation rejection")
	}

	allowed, err := policy.Enforce(cozoapi.PreparedQuery{
		Script: "?[x] := *users{x}",
		Opts: &cozoapi.QueryOptions{
			Limit:      cozoapi.PtrInt64(100),
			TimeoutSec: cozoapi.PtrFloat64(5),
		},
		Meta: cozoapi.QueryMeta{Relations: []string{"users"}},
	})
	if err != nil {
		t.Fatalf("expected query allowed: %v", err)
	}
	if *allowed.Opts.Limit != 10 || *allowed.Opts.TimeoutSec != 2 {
		t.Fatalf("normalized opts = %#v", allowed.Opts)
	}
}

func TestDBAndRelationHandle(t *testing.T) {
	backend := fakebackend.New()
	backend.PushExecResult(cozoapi.CozoResult{
		Headers: []string{"id"},
		Rows:    []cozoapi.CozoRow{{"u1"}},
	}, nil)
	backend.PushExecResult(cozoapi.CozoResult{
		Headers: []string{"id", "name"},
		Rows:    []cozoapi.CozoRow{{"u1", "Ada"}},
	}, nil)

	db, err := cozoapi.Open(backend, cozoapi.DefaultPolicy())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	result, err := db.ExecScript(context.Background(), "?[id] <- [[1]]", nil, nil)
	if err != nil {
		t.Fatalf("exec script: %v", err)
	}
	if result.Scalar() != "u1" {
		t.Fatalf("scalar = %#v", result.Scalar())
	}

	row, err := db.Rel("users").Get(context.Background(), map[string]cozoapi.CozoValue{"id": "u1"})
	if err != nil {
		t.Fatalf("relation get: %v", err)
	}
	if row["name"] != "Ada" {
		t.Fatalf("row = %#v", row)
	}

	calls := backend.Calls()
	if len(calls) != 2 {
		t.Fatalf("calls = %d", len(calls))
	}

	if err := db.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
	if !backend.Closed() {
		t.Fatalf("backend must be closed")
	}
	if _, err := db.ExecScript(context.Background(), "?[x] <- [[1]]", nil, nil); err == nil {
		t.Fatalf("expected closed error")
	}
}
