package module

import (
	"context"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/fakebackend"
)

func TestModuleOpenAndExecFlow(t *testing.T) {
	backend := fakebackend.New()
	db, err := cozoapi.Open(backend, cozoapi.Policy{
		AllowSystemOps: true,
		QueryOptions: cozoapi.QueryOptionPolicy{
			MaxLimit: cozoapi.PtrInt64(10),
		},
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	mod := New(func(_ context.Context, _ OpenOptions) (*cozoapi.DB, error) {
		return db, nil
	})
	vm, req := newRuntimeWithModule(t, mod)

	cozodb := requireModule(t, vm, req)
	dbValue := call(t, cozodb.Get("open"), cozodb)
	dbObj := dbValue.ToObject(vm)

	// exec(script, params, opts)
	opts := vm.NewObject()
	_ = opts.Set("limit", int64(20))
	call(t, dbObj.Get("exec"), dbValue, vm.ToValue("?[id] <- [[$p1]]"), vm.ToValue(map[string]any{"p1": "u1"}), opts)
	if len(backend.Calls()) != 1 {
		t.Fatalf("calls = %d, want 1", len(backend.Calls()))
	}
	if backend.Calls()[0].Opts.Limit == nil || *backend.Calls()[0].Opts.Limit != 10 {
		t.Fatalf("expected limit clamped to 10, got %#v", backend.Calls()[0].Opts)
	}

	// q(opts)`...`
	qWithOpts := call(t, dbObj.Get("q"), dbValue, vm.ToValue(map[string]any{"limit": int64(99)}))
	template := vm.NewArray(vm.ToValue("?[id] <- [["), vm.ToValue("]]"))
	call(t, qWithOpts, goja.Undefined(), template, vm.ToValue("u2"))
	if len(backend.Calls()) != 2 {
		t.Fatalf("calls = %d, want 2", len(backend.Calls()))
	}
	if !strings.Contains(backend.Calls()[1].Script, "$p1") {
		t.Fatalf("q compiled script = %q", backend.Calls()[1].Script)
	}
	if backend.Calls()[1].Opts.Limit == nil || *backend.Calls()[1].Opts.Limit != 10 {
		t.Fatalf("expected q limit clamped to 10, got %#v", backend.Calls()[1].Opts)
	}

	// cq(opts)`...` + atomic([...])
	cqWithOpts := call(t, dbObj.Get("cq"), dbValue, vm.ToValue(map[string]any{"limit": int64(7)}))
	prepared := call(t, cqWithOpts, goja.Undefined(), template, vm.ToValue("u3"))
	atomicQueries := vm.NewArray(prepared)
	call(t, dbObj.Get("atomic"), dbValue, atomicQueries)
	if len(backend.Calls()) != 3 {
		t.Fatalf("calls = %d, want 3", len(backend.Calls()))
	}
	if !strings.Contains(backend.Calls()[2].Script, "{") || !strings.Contains(backend.Calls()[2].Script, "$q1_p1") {
		t.Fatalf("atomic script = %q", backend.Calls()[2].Script)
	}

	// rel(name).put(...)
	rel := call(t, dbObj.Get("rel"), dbValue, vm.ToValue("users"))
	relObj := rel.ToObject(vm)
	call(
		t,
		relObj.Get("put"),
		rel,
		vm.ToValue([]map[string]any{{"id": "u9", "name": "Ada"}}),
		vm.ToValue(map[string]any{"returning": true}),
	)
	if len(backend.Calls()) != 4 {
		t.Fatalf("calls = %d, want 4", len(backend.Calls()))
	}
	if !strings.Contains(backend.Calls()[3].Script, ":put users") || !strings.Contains(backend.Calls()[3].Script, ":returning") {
		t.Fatalf("put script = %q", backend.Calls()[3].Script)
	}
}

func TestModuleRelSystemOpsRespectPolicy(t *testing.T) {
	backend := fakebackend.New()
	db, err := cozoapi.Open(backend, cozoapi.Policy{
		AllowSystemOps: false,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	mod := New(func(_ context.Context, _ OpenOptions) (*cozoapi.DB, error) {
		return db, nil
	})
	vm, req := newRuntimeWithModule(t, mod)

	cozodb := requireModule(t, vm, req)
	dbValue := call(t, cozodb.Get("open"), cozodb)
	rel := call(t, dbValue.ToObject(vm).Get("rel"), dbValue, vm.ToValue("users"))
	rejected := callRejected(t, rel.ToObject(vm).Get("columns"), rel)
	if !strings.Contains(rejected, "policy rejected system operation") {
		t.Fatalf("unexpected rejection message: %q", rejected)
	}

	if len(backend.Calls()) != 0 {
		t.Fatalf("columns should be rejected by policy before backend call")
	}
}

func newRuntimeWithModule(t *testing.T, mod *Module) (*goja.Runtime, *require.RequireModule) {
	t.Helper()
	reg := require.NewRegistry()
	mod.Register(reg)
	vm := goja.New()
	req := reg.Enable(vm)
	return vm, req
}

func requireModule(t *testing.T, vm *goja.Runtime, req *require.RequireModule) *goja.Object {
	t.Helper()
	value, err := req.Require("cozodb")
	if err != nil {
		t.Fatalf("require cozodb: %v", err)
	}
	return value.ToObject(vm)
}

func call(t *testing.T, fnValue goja.Value, this goja.Value, args ...goja.Value) goja.Value {
	t.Helper()
	fn, ok := goja.AssertFunction(fnValue)
	if !ok {
		t.Fatalf("value is not callable: %s", fnValue.String())
	}
	v, err := fn(this, args...)
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if p, ok := v.Export().(*goja.Promise); ok {
		switch p.State() {
		case goja.PromiseStateRejected:
			t.Fatalf("promise rejected: %v", p.Result().Export())
		case goja.PromiseStateFulfilled:
			return p.Result()
		default:
			t.Fatalf("promise still pending")
		}
	}
	return v
}

func callRejected(t *testing.T, fnValue goja.Value, this goja.Value, args ...goja.Value) string {
	t.Helper()
	fn, ok := goja.AssertFunction(fnValue)
	if !ok {
		t.Fatalf("value is not callable: %s", fnValue.String())
	}
	v, err := fn(this, args...)
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}
	p, ok := v.Export().(*goja.Promise)
	if !ok {
		t.Fatalf("expected promise return")
	}
	if p.State() != goja.PromiseStateRejected {
		t.Fatalf("expected rejected promise, got state %v", p.State())
	}
	result := p.Result().Export()
	if asMap, ok := result.(map[string]any); ok {
		if msg, ok := asMap["message"].(string); ok {
			return msg
		}
	}
	return ""
}
