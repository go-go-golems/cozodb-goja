//go:build ignore
// +build ignore

// Bypass cozoapi entirely — call cozo-lib-go directly to isolate the bug.
//
// Run:
//   cd cozodb-goja
//   CGO_LDFLAGS="-L$PWD/.deps/cozo" GOWORK=off go run -tags cozo_cgo \
//     ./ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/02-cozo-raw-cgo-repro.go

package main

import (
	"fmt"
	"os"

	cozo "github.com/cozodb/cozo-lib-go"
)

func main() {
	fmt.Println("=== Test: Raw cozo-lib-go, mem engine ===")
	db, err := cozo.New("mem", "", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open mem: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("  opened mem db OK")

	fmt.Println("\n--- :create ---")
	res, err := db.Run(`:create test {id: String => val: String}`, nil)
	fmt.Printf("  ok=%v err=%v\n", res.Ok, err)

	fmt.Println("\n--- :put ---")
	res, err = db.Run(`?[id, val] <- [["a", "hello"]]
:put test {id => val}`, nil)
	fmt.Printf("  ok=%v err=%v\n", res.Ok, err)

	fmt.Println("\n--- read ---")
	res, err = db.Run(`?[id, val] := *test{id, val}`, nil)
	fmt.Printf("  ok=%v err=%v headers=%v rows=%v\n", res.Ok, err, res.Headers, res.Rows)

	fmt.Println("\n=== Test: Raw cozo-lib-go, sqlite engine ===")
	dbPath := "/tmp/cozo-repro-test.db"
	os.Remove(dbPath)
	db2, err := cozo.New("sqlite", dbPath, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open sqlite: %v\n", err)
		os.Exit(1)
	}
	defer db2.Close()
	defer os.Remove(dbPath)
	fmt.Println("  opened sqlite db OK")

	fmt.Println("\n--- :create ---")
	res, err = db2.Run(`:create test {id: String => val: String}`, nil)
	fmt.Printf("  ok=%v err=%v\n", res.Ok, err)

	fmt.Println("\n--- :put ---")
	res, err = db2.Run(`?[id, val] <- [["a", "hello"]]
:put test {id => val}`, nil)
	fmt.Printf("  ok=%v err=%v\n", res.Ok, err)

	fmt.Println("\n--- read ---")
	res, err = db2.Run(`?[id, val] := *test{id, val}`, nil)
	fmt.Printf("  ok=%v err=%v headers=%v rows=%v\n", res.Ok, err, res.Headers, res.Rows)
}
