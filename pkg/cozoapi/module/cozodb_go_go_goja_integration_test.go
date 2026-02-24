package module

import (
	"context"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/fakebackend"
	ggjengine "github.com/go-go-golems/go-go-goja/engine"
)

func TestModuleWithGoGoGojaRuntime(t *testing.T) {
	backend := fakebackend.New()
	db, err := cozoapi.Open(backend, cozoapi.DefaultPolicy())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	mod := New(func(_ context.Context, _ OpenOptions) (*cozoapi.DB, error) {
		return db, nil
	})

	factory, err := ggjengine.NewBuilder().
		WithModules(
			ggjengine.DefaultRegistryModules(),
			ggjengine.NativeModuleSpec{
				ModuleID:   "cozodb-test-module",
				ModuleName: mod.Name(),
				Loader:     mod.Loader,
			},
		).
		Build()
	if err != nil {
		t.Fatalf("build runtime factory: %v", err)
	}

	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer func() {
		_ = rt.Close(context.Background())
	}()

	_, err = rt.Owner.Call(context.Background(), "cozodb.integration", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, err := vm.RunString(`
      const cozo = require("cozodb");
      const db = cozo.open();
      db.exec("?[id, name] <- [[$id, $name]]", { id: "u1", name: "Ada" });
      db.rel("users").put([{ id: "u1", name: "Ada" }], { returning: true });
    `)
		return nil, err
	})
	if err != nil {
		t.Fatalf("execute integration script: %v", err)
	}

	calls := backend.Calls()
	if len(calls) != 2 {
		t.Fatalf("backend calls = %d, want 2", len(calls))
	}
}
