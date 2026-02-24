package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/module"
)

func main() {
	opts := parseFlags()
	vm := newRuntime()

	switch {
	case opts.eval != "":
		if err := runSnippet(vm, "<eval>", opts.eval); err != nil {
			exitWithError(err)
		}
		return
	case opts.script != "":
		data, err := os.ReadFile(opts.script)
		if err != nil {
			exitWithError(fmt.Errorf("read script %q: %w", opts.script, err))
		}
		if err := runSnippet(vm, opts.script, string(data)); err != nil {
			exitWithError(err)
		}
		return
	default:
		if err := runREPL(vm); err != nil {
			exitWithError(err)
		}
	}
}

type cliOptions struct {
	script string
	eval   string
}

func parseFlags() cliOptions {
	var opts cliOptions
	flag.StringVar(&opts.script, "script", "", "Path to JavaScript file to execute")
	flag.StringVar(&opts.eval, "eval", "", "Inline JavaScript snippet to execute")
	flag.Parse()
	return opts
}

func newRuntime() *goja.Runtime {
	vm := goja.New()
	reg := require.NewRegistry()
	module.New(module.DefaultOpen).Register(reg)
	reg.Enable(vm)
	return vm
}

func runSnippet(vm *goja.Runtime, sourceName, code string) error {
	value, err := vm.RunScript(sourceName, code)
	if err != nil {
		return fmt.Errorf("javascript execution failed: %w", err)
	}
	fmt.Println(renderValue(value))
	return nil
}

func runREPL(vm *goja.Runtime) error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprintln(os.Stdout, "cozodb-goja repl (type exit to quit)")
	for {
		fmt.Fprint(os.Stdout, "js> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read repl input: %w", err)
			}
			fmt.Fprintln(os.Stdout)
			return nil
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == ".exit" || line == "quit" {
			return nil
		}
		if err := runSnippet(vm, "<repl>", line); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}

func renderValue(v goja.Value) string {
	if v == nil || goja.IsUndefined(v) {
		return "undefined"
	}
	if p, ok := v.Export().(*goja.Promise); ok {
		switch p.State() {
		case goja.PromiseStateFulfilled:
			return fmt.Sprintf("Promise<fulfilled>: %v", p.Result().Export())
		case goja.PromiseStateRejected:
			return fmt.Sprintf("Promise<rejected>: %v", p.Result().Export())
		default:
			return "Promise<pending>"
		}
	}
	return fmt.Sprintf("%v", v.Export())
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
