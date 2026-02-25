// Direct C-API repro that uses the 4-argument cozo_run_query signature.
//
// Usage:
//   CGO_LDFLAGS="-L$PWD/.deps/cozo" GOWORK=off go run -tags cozo_cgo \
//     ./ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/03-cozo-c-api-4arg-repro.go
package main

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

/*
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#cgo LDFLAGS: -lcozo_c -lstdc++ -lm

char *cozo_open_db(const char *engine, const char *path, const char *options, int32_t *db_id);
bool cozo_close_db(int32_t id);
char *cozo_run_query(int32_t db_id, const char *script_raw, const char *params_raw, bool immutable_query);
void cozo_free_str(char *s);
*/
import "C"

type namedRows struct {
	Headers []string `json:"headers"`
	Rows    [][]any  `json:"rows"`
	Ok      bool     `json:"ok"`
}

func main() {
	dbID, err := openDB("mem", "", "{}")
	if err != nil {
		panic(err)
	}
	defer C.cozo_close_db(dbID)

	run(dbID, ":create t {id: String => val: String}", false)
	run(dbID, "?[id, val] <- [[\"a\", \"x\"]]\n:put t {id => val}", false)
	run(dbID, "?[id, val] := *t{id, val}", true)
}

func openDB(engine, path, options string) (C.int32_t, error) {
	cEngine := C.CString(engine)
	defer C.free(unsafe.Pointer(cEngine))
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	cOpts := C.CString(options)
	defer C.free(unsafe.Pointer(cOpts))

	var dbID C.int32_t
	errPtr := C.cozo_open_db(cEngine, cPath, cOpts, &dbID)
	if errPtr != nil {
		defer C.cozo_free_str(errPtr)
		return 0, fmt.Errorf("cozo_open_db: %s", C.GoString(errPtr))
	}
	return dbID, nil
}

func run(dbID C.int32_t, script string, immutable bool) {
	cScript := C.CString(script)
	defer C.free(unsafe.Pointer(cScript))
	cParams := C.CString("{}")
	defer C.free(unsafe.Pointer(cParams))

	resPtr := C.cozo_run_query(dbID, cScript, cParams, C.bool(immutable))
	if resPtr == nil {
		fmt.Printf("script: %s\nresult: <nil>\n\n", script)
		return
	}
	defer C.cozo_free_str(resPtr)

	raw := C.GoString(resPtr)
	var out namedRows
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		fmt.Printf("script: %s\nraw: %s\njson err: %v\n\n", script, raw, err)
		return
	}
	fmt.Printf("script: %s\nok=%v headers=%v rows=%v\n\n", script, out.Ok, out.Headers, out.Rows)
}
