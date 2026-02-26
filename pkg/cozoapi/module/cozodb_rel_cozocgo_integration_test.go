//go:build cozo_cgo && cozo_cgo_integration

package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
)

func TestRelLifecycleWithCozoCGOBackend(t *testing.T) {
	t.Parallel()

	mod := New(DefaultOpen)
	vm, req := newRuntimeWithModule(t, mod)

	cozodb := requireModule(t, vm, req)
	dbPath := filepath.Join(t.TempDir(), "rel-polish.sqlite")
	dbValue := call(
		t,
		cozodb.Get("open"),
		cozodb,
		vm.ToValue(map[string]any{
			"backend": "cozo_cgo",
			"engine":  "sqlite",
			"path":    dbPath,
		}),
	)
	dbObj := dbValue.ToObject(vm)
	relValue := call(t, dbObj.Get("rel"), dbValue, vm.ToValue("users"))
	relObj := relValue.ToObject(vm)

	call(
		t,
		relObj.Get("create"),
		relValue,
		vm.ToValue(map[string]any{
			"keys": map[string]any{
				"id": "String",
			},
			"values": map[string]any{
				"name": "String",
				"team": "String",
			},
		}),
	)

	call(
		t,
		relObj.Get("put"),
		relValue,
		vm.ToValue(map[string]any{
			"headers": []string{"id", "name", "team"},
			"rows": [][]any{
				{"u1", "Ada", "db"},
				{"u2", "Bob", "platform"},
			},
		}),
		vm.ToValue(map[string]any{"returning": true}),
	)

	call(
		t,
		relObj.Get("insert"),
		relValue,
		vm.ToValue([]map[string]any{
			{"id": "u3", "name": "Cleo", "team": "ml"},
		}),
	)

	call(
		t,
		relObj.Get("update"),
		relValue,
		vm.ToValue([]map[string]any{
			{"id": "u1", "name": "Ada Lovelace", "team": "db"},
		}),
	)

	gotU1 := call(t, relObj.Get("get"), relValue, vm.ToValue(map[string]any{"id": "u1"}))
	gotU1Map := map[string]any{}
	if err := vm.ExportTo(gotU1, &gotU1Map); err != nil {
		t.Fatalf("export get(u1) result: %v", err)
	}
	if gotU1Map["id"] != "u1" {
		t.Fatalf("expected id=u1, got %#v", gotU1Map["id"])
	}
	if gotU1Map["name"] != "Ada Lovelace" {
		t.Fatalf("expected updated name, got %#v", gotU1Map["name"])
	}

	call(
		t,
		relObj.Get("rm"),
		relValue,
		vm.ToValue([]map[string]any{
			{"id": "u2"},
		}),
	)
	call(
		t,
		relObj.Get("del"),
		relValue,
		vm.ToValue([]map[string]any{
			{"id": "u3", "name": "Cleo", "team": "ml"},
		}),
	)

	gotU2 := call(t, relObj.Get("get"), relValue, vm.ToValue(map[string]any{"id": "u2"}))
	if !goja.IsUndefined(gotU2) {
		t.Fatalf("expected get(u2) to return undefined after rm, got %v", gotU2.Export())
	}

	gotU3 := call(t, relObj.Get("get"), relValue, vm.ToValue(map[string]any{"id": "u3"}))
	if !goja.IsUndefined(gotU3) {
		t.Fatalf("expected get(u3) to return undefined after del, got %v", gotU3.Export())
	}

	columnsResult := call(t, relObj.Get("columns"), relValue)
	columnsObj := columnsResult.ToObject(vm)
	var headers []string
	if err := vm.ExportTo(columnsObj.Get("headers"), &headers); err != nil {
		t.Fatalf("export columns headers: %v", err)
	}
	if len(headers) == 0 {
		t.Fatalf("expected columns headers to be non-empty")
	}

	_ = call(t, relObj.Get("indices"), relValue)
	_ = call(t, relObj.Get("access"), relValue, vm.ToValue("normal"))

	_ = call(t, dbObj.Get("close"), dbValue)
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected sqlite file to exist at %s: %v", dbPath, err)
	}
}
