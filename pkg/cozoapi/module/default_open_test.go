package module

import (
	"context"
	"testing"
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
