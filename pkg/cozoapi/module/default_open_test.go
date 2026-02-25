package module

import (
	"context"
	"strings"
	"testing"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/cozocgo"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/fakebackend"
)

func TestDefaultOpenFake(t *testing.T) {
	db, err := DefaultOpen(context.Background(), OpenOptions{})
	if err != nil {
		t.Fatalf("default open failed: %v", err)
	}
	if db.Backend() != "fake" {
		t.Fatalf("backend = %q, want fake", db.Backend())
	}
	if err := db.Close(context.Background()); err != nil {
		t.Fatalf("close db: %v", err)
	}
}

func TestDefaultOpenUnknownBackend(t *testing.T) {
	_, err := DefaultOpen(context.Background(), OpenOptions{Backend: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported backend error")
	}
}

func TestDefaultOpenCozocgoForwardsOptions(t *testing.T) {
	old := openCozoBackend
	defer func() {
		openCozoBackend = old
	}()

	var captured cozocgo.OpenOptions
	openCozoBackend = func(_ context.Context, opts cozocgo.OpenOptions) (cozoapi.Backend, error) {
		captured = opts
		return fakebackend.New(), nil
	}

	db, err := DefaultOpen(context.Background(), OpenOptions{
		Backend: "cozocgo",
		Engine:  "sqlite",
		Path:    "/tmp/cozo.db",
		Options: map[string]any{
			"create": true,
		},
	})
	if err != nil {
		t.Fatalf("default open cozocgo failed: %v", err)
	}
	if db.Backend() != "fake" {
		t.Fatalf("backend = %q, want fake from stub backend", db.Backend())
	}

	if captured.Engine != "sqlite" {
		t.Fatalf("engine = %q, want sqlite", captured.Engine)
	}
	if captured.Path != "/tmp/cozo.db" {
		t.Fatalf("path = %q", captured.Path)
	}
	create, ok := captured.Options["create"].(bool)
	if !ok || !create {
		t.Fatalf("options.create = %#v", captured.Options["create"])
	}
}

func TestDefaultOpenCozocgoDefaultsEngine(t *testing.T) {
	old := openCozoBackend
	defer func() {
		openCozoBackend = old
	}()

	var captured cozocgo.OpenOptions
	openCozoBackend = func(_ context.Context, opts cozocgo.OpenOptions) (cozoapi.Backend, error) {
		captured = opts
		return nil, context.Canceled
	}

	_, err := DefaultOpen(context.Background(), OpenOptions{Backend: "cozo_cgo"})
	if err == nil {
		t.Fatalf("expected error from stubbed backend opener")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("unexpected error = %v", err)
	}
	if captured.Engine != "mem" {
		t.Fatalf("engine = %q, want mem", captured.Engine)
	}
}
