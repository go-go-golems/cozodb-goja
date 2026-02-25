// Minimal reproduction: CozoDB `:create` fails with "write lock required for read-only query"
// Tests 4 paths to isolate whether the issue is :timeout, the DB wrapper, or the backend.
//
// Run:
//   cd cozodb-goja
//   CGO_LDFLAGS="-L$PWD/.deps/cozo" GOWORK=off go run -tags cozo_cgo \
//     ./ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/01-cozo-write-lock-repro.go
//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/cozocgo"
)

func main() {
	ctx := context.Background()

	// 1. Direct backend, no policy wrapper, zero opts
	backend1, err := cozocgo.Open(ctx, cozocgo.OpenOptions{Engine: "mem"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Test 1: Direct backend.Exec, zero QueryOptions ===")
	_, err = backend1.Exec(ctx, `:create test1 {id: String => val: String}`, nil, cozoapi.QueryOptions{})
	fmt.Printf("  result: %v\n", err)

	// 2. Direct backend, with :timeout appended manually
	fmt.Println("\n=== Test 2: Direct backend.Exec, :timeout 3 appended ===")
	_, err = backend1.Exec(ctx, ":create test2 {id: String => val: String}\n:timeout 3", nil, cozoapi.QueryOptions{})
	fmt.Printf("  result: %v\n", err)

	// 3. Through DB wrapper, DefaultPolicy (no timeout in policy)
	fmt.Println("\n=== Test 3: DB wrapper, DefaultPolicy, no timeout ===")
	policy3 := cozoapi.DefaultPolicy()
	policy3.AllowSystemOps = true

	backend3, _ := cozocgo.Open(ctx, cozocgo.OpenOptions{Engine: "mem"})
	db3, _ := cozoapi.Open(backend3, policy3)
	_, err = db3.ExecScript(ctx, `:create test3 {id: String => val: String}`, nil, nil)
	fmt.Printf("  result: %v\n", err)

	// 4. Through DB wrapper, policy with DefaultTimeoutSec=3
	fmt.Println("\n=== Test 4: DB wrapper, DefaultTimeoutSec=3 ===")
	policy4 := cozoapi.DefaultPolicy()
	policy4.AllowSystemOps = true
	policy4.QueryOptions.DefaultTimeoutSec = cozoapi.PtrFloat64(3)

	backend4, _ := cozocgo.Open(ctx, cozocgo.OpenOptions{Engine: "mem"})
	db4, _ := cozoapi.Open(backend4, policy4)
	_, err = db4.ExecScript(ctx, `:create test4 {id: String => val: String}`, nil, nil)
	fmt.Printf("  result: %v\n", err)

	// 5. Through DB wrapper with timeout, but passing explicit opts with no timeout
	fmt.Println("\n=== Test 5: DB wrapper with timeout policy, explicit empty opts ===")
	backend5, _ := cozocgo.Open(ctx, cozocgo.OpenOptions{Engine: "mem"})
	db5, _ := cozoapi.Open(backend5, policy4)
	noTimeout := &cozoapi.QueryOptions{TimeoutSec: cozoapi.PtrFloat64(0)}
	_, err = db5.ExecScript(ctx, `:create test5 {id: String => val: String}`, nil, noTimeout)
	fmt.Printf("  result: %v\n", err)

	// 6. Direct backend with :put (needs existing relation)
	fmt.Println("\n=== Test 6: :put after successful :create (if test 1 passed) ===")
	_, err = backend1.Exec(ctx, `?[id, val] <- [["a", "hello"]]
:put test1 {id => val}`, nil, cozoapi.QueryOptions{})
	fmt.Printf("  result: %v\n", err)

	// 7. Query the data back
	fmt.Println("\n=== Test 7: Read query ===")
	res, err := backend1.Exec(ctx, `?[id, val] := *test1{id, val}`, nil, cozoapi.QueryOptions{})
	if err != nil {
		fmt.Printf("  result: %v\n", err)
	} else {
		fmt.Printf("  headers: %v\n", res.Headers)
		fmt.Printf("  rows:    %v\n", res.Rows)
	}
}
