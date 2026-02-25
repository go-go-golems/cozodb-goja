//go:build ignore
// +build ignore

// Validate that the tagged cozocgo adapter can export and import relation rows.
//
// Run:
//
//	cd cozodb-goja
//	CGO_LDFLAGS="-L$PWD/.deps/cozo" GOWORK=off go run -tags cozo_cgo \
//	  ./ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/scripts/03-cozocgo-export-import-repro.go
package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/cozocgo"
)

func main() {
	ctx := context.Background()
	b, err := cozocgo.Open(ctx, cozocgo.OpenOptions{Engine: "mem"})
	if err != nil {
		panic(err)
	}
	defer func() { _ = b.Close(ctx) }()

	if _, err := b.Exec(ctx, `:create t {id: String => val: String}`, nil, cozoapi.QueryOptions{}); err != nil {
		panic(err)
	}
	if _, err := b.Exec(ctx, `?[id, val] <- [["a", "hello"]] :put t {id => val}`, nil, cozoapi.QueryOptions{}); err != nil {
		panic(err)
	}

	exported, err := b.Export(ctx, []string{"t"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("exported relation t: headers=%v rows=%v\n", exported["t"].Headers, exported["t"].Rows)

	if err := b.Import(ctx, map[string]cozoapi.RelationRows{
		"t": {
			Headers: []string{"id", "val"},
			Rows:    []cozoapi.CozoRow{{"b", "bye"}},
		},
	}); err != nil {
		panic(err)
	}

	res, err := b.Exec(ctx, `?[id, val] := *t{id,val}`, nil, cozoapi.QueryOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("query rows after import: %v\n", res.Rows)
}
