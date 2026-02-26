package module

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

type OpenOptions struct {
	Backend string         `json:"backend"`
	Engine  string         `json:"engine"`
	Path    string         `json:"path"`
	Options map[string]any `json:"options"`
}

type OpenFunc func(context.Context, OpenOptions) (*cozoapi.DB, error)

type Module struct {
	open OpenFunc
}

func New(open OpenFunc) *Module {
	return &Module{open: open}
}

func (m *Module) Name() string {
	return "cozodb"
}

func (m *Module) Doc() string {
	return "CozoDB module exposing open()/exec()/q()/cq()/atomic()/rel()/export()/import()/close() over a backend adapter."
}

func (m *Module) Register(reg *require.Registry) {
	reg.RegisterNativeModule(m.Name(), m.Loader)
}

func (m *Module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").ToObject(vm)
	_ = exports.Set("open", func(call goja.FunctionCall) goja.Value {
		if m.open == nil {
			panic(vm.ToValue("cozodb module has no open function configured"))
		}

		opts := OpenOptions{Backend: "fake"}
		if len(call.Arguments) > 0 {
			decoded, err := decodeOpenOptions(vm, call.Arguments[0])
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			opts = decoded
		}

		db, err := m.open(context.Background(), opts)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return m.newDBHandle(vm, db)
	})
}

func (m *Module) newDBHandle(vm *goja.Runtime, db *cozoapi.DB) goja.Value {
	handle := vm.NewObject()
	_ = handle.Set("backend", db.Backend())
	_ = handle.Set("capabilities", db.Capabilities())

	_ = handle.Set("exec", func(call goja.FunctionCall) goja.Value {
		return promise(vm, func() (goja.Value, error) {
			query, err := decodeExecQuery(vm, call.Arguments)
			if err != nil {
				return nil, err
			}
			result, err := db.ExecPrepared(context.Background(), query)
			if err != nil {
				return nil, err
			}
			return resultToValue(vm, result), nil
		})
	})

	_ = handle.Set("q", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(vm.ToValue("q requires template strings or options"))
		}
		if parts, ok, err := decodeTemplateParts(vm, call.Arguments[0]); err != nil {
			panic(vm.ToValue(err.Error()))
		} else if ok {
			values := make([]cozoapi.CozoValue, 0, len(call.Arguments)-1)
			for _, arg := range call.Arguments[1:] {
				values = append(values, arg.Export())
			}
			return promise(vm, func() (goja.Value, error) {
				result, err := db.Q(context.Background(), parts, values, nil)
				if err != nil {
					return nil, err
				}
				return resultToValue(vm, result), nil
			})
		}

		opts, err := decodeQueryOptions(vm, call.Arguments[0])
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}

		return vm.ToValue(func(inner goja.FunctionCall) goja.Value {
			if len(inner.Arguments) == 0 {
				panic(vm.ToValue("q(opts) requires template strings"))
			}
			parts, ok, err := decodeTemplateParts(vm, inner.Arguments[0])
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			if !ok {
				panic(vm.ToValue("q(opts) expects template strings as first argument"))
			}
			values := make([]cozoapi.CozoValue, 0, len(inner.Arguments)-1)
			for _, arg := range inner.Arguments[1:] {
				values = append(values, arg.Export())
			}
			return promise(vm, func() (goja.Value, error) {
				result, err := db.Q(context.Background(), parts, values, opts)
				if err != nil {
					return nil, err
				}
				return resultToValue(vm, result), nil
			})
		})
	})

	_ = handle.Set("cq", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(vm.ToValue("cq requires template strings or options"))
		}
		if parts, ok, err := decodeTemplateParts(vm, call.Arguments[0]); err != nil {
			panic(vm.ToValue(err.Error()))
		} else if ok {
			values := make([]cozoapi.CozoValue, 0, len(call.Arguments)-1)
			for _, arg := range call.Arguments[1:] {
				values = append(values, arg.Export())
			}
			query, err := db.CQ(parts, values, nil)
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			return preparedQueryToValue(vm, query)
		}

		opts, err := decodeQueryOptions(vm, call.Arguments[0])
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(func(inner goja.FunctionCall) goja.Value {
			if len(inner.Arguments) == 0 {
				panic(vm.ToValue("cq(opts) requires template strings"))
			}
			parts, ok, err := decodeTemplateParts(vm, inner.Arguments[0])
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			if !ok {
				panic(vm.ToValue("cq(opts) expects template strings as first argument"))
			}
			values := make([]cozoapi.CozoValue, 0, len(inner.Arguments)-1)
			for _, arg := range inner.Arguments[1:] {
				values = append(values, arg.Export())
			}
			query, err := db.CQ(parts, values, opts)
			if err != nil {
				panic(vm.ToValue(err.Error()))
			}
			return preparedQueryToValue(vm, query)
		})
	})

	_ = handle.Set("atomic", func(call goja.FunctionCall) goja.Value {
		return promise(vm, func() (goja.Value, error) {
			queries, opts, err := decodeAtomicInput(vm, call.Arguments)
			if err != nil {
				return nil, err
			}
			result, err := db.Atomic(context.Background(), queries, opts)
			if err != nil {
				return nil, err
			}
			return resultToValue(vm, result), nil
		})
	})

	_ = handle.Set("rel", func(name string) goja.Value {
		rel := db.Rel(name)
		return relationHandleToValue(vm, rel)
	})

	_ = handle.Set("export", func(call goja.FunctionCall) goja.Value {
		return promise(vm, func() (goja.Value, error) {
			relations, err := decodeStringArray(vm, call.Arguments)
			if err != nil {
				return nil, err
			}
			data, err := db.Export(context.Background(), relations)
			if err != nil {
				return nil, err
			}
			return vm.ToValue(data), nil
		})
	})

	_ = handle.Set("import", func(call goja.FunctionCall) goja.Value {
		return promise(vm, func() (goja.Value, error) {
			if len(call.Arguments) < 1 {
				return nil, fmt.Errorf("import requires data object")
			}
			data := map[string]cozoapi.RelationRows{}
			if err := vm.ExportTo(call.Arguments[0], &data); err != nil {
				return nil, fmt.Errorf("decode import data: %w", err)
			}
			if err := db.Import(context.Background(), data); err != nil {
				return nil, err
			}
			return goja.Undefined(), nil
		})
	})

	_ = handle.Set("close", func() goja.Value {
		return promise(vm, func() (goja.Value, error) {
			if err := db.Close(context.Background()); err != nil {
				return nil, err
			}
			return goja.Undefined(), nil
		})
	})

	return handle
}

func relationHandleToValue(vm *goja.Runtime, rel *cozoapi.RelationHandle) goja.Value {
	obj := vm.NewObject()
	_ = obj.Set("name", rel.Name())

	_ = obj.Set("create", func(call goja.FunctionCall) goja.Value {
		return relationPromise(vm, "rel.create", func() (goja.Value, error) {
			if len(call.Arguments) < 1 {
				return nil, fmt.Errorf("rel.create requires a spec object")
			}
			spec, err := decodeRelationSpec(vm, call.Arguments[0])
			if err != nil {
				return nil, err
			}
			opts, err := decodeRelationCreateOptions(vm, argAt(call.Arguments, 1))
			if err != nil {
				return nil, err
			}
			if err := rel.Create(context.Background(), spec, opts); err != nil {
				return nil, err
			}
			return goja.Undefined(), nil
		})
	})

	buildMutation := func(
		operation string,
		run func(ctx context.Context, rows []map[string]cozoapi.CozoValue, opts cozoapi.RelationMutationOptions) (cozoapi.CozoResult, error),
	) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			return relationPromise(vm, operation, func() (goja.Value, error) {
				rows, opts, err := decodeMutationInput(vm, call.Arguments)
				if err != nil {
					return nil, err
				}
				result, err := run(context.Background(), rows, opts)
				if err != nil {
					return nil, err
				}
				return resultToValue(vm, result), nil
			})
		}
	}

	_ = obj.Set("put", buildMutation("rel.put", rel.Put))
	_ = obj.Set("insert", buildMutation("rel.insert", rel.Insert))
	_ = obj.Set("update", buildMutation("rel.update", rel.Update))
	_ = obj.Set("rm", buildMutation("rel.rm", rel.Rm))
	_ = obj.Set("del", buildMutation("rel.del", rel.Del))

	_ = obj.Set("get", func(call goja.FunctionCall) goja.Value {
		return relationPromise(vm, "rel.get", func() (goja.Value, error) {
			if len(call.Arguments) == 0 {
				return nil, fmt.Errorf("rel.get requires key object")
			}
			key := map[string]cozoapi.CozoValue{}
			if err := vm.ExportTo(call.Arguments[0], &key); err != nil {
				return nil, fmt.Errorf("decode relation key: %w", err)
			}
			row, err := rel.Get(context.Background(), key)
			if err != nil {
				return nil, err
			}
			if row == nil {
				return goja.Undefined(), nil
			}
			return vm.ToValue(row), nil
		})
	})

	_ = obj.Set("columns", func() goja.Value {
		return relationPromise(vm, "rel.columns", func() (goja.Value, error) {
			result, err := rel.Columns(context.Background())
			if err != nil {
				return nil, err
			}
			return resultToValue(vm, result), nil
		})
	})

	_ = obj.Set("indices", func() goja.Value {
		return relationPromise(vm, "rel.indices", func() (goja.Value, error) {
			result, err := rel.Indices(context.Background())
			if err != nil {
				return nil, err
			}
			return resultToValue(vm, result), nil
		})
	})

	_ = obj.Set("access", func(level string) goja.Value {
		return relationPromise(vm, "rel.access", func() (goja.Value, error) {
			if err := rel.Access(context.Background(), cozoapi.AccessLevel(level)); err != nil {
				return nil, err
			}
			return goja.Undefined(), nil
		})
	})

	return obj
}

func promise(vm *goja.Runtime, fn func() (goja.Value, error)) goja.Value {
	return promiseWithError(vm, "COZO_JS_ERROR", "", fn)
}

func relationPromise(vm *goja.Runtime, operation string, fn func() (goja.Value, error)) goja.Value {
	return promiseWithError(vm, "COZO_REL_ERROR", operation, fn)
}

func promiseWithError(vm *goja.Runtime, code, operation string, fn func() (goja.Value, error)) goja.Value {
	p, resolve, reject := vm.NewPromise()
	result, err := fn()
	if err != nil {
		errObj := errorObject(vm, code, operation, err)
		_ = reject(errObj)
		return vm.ToValue(p)
	}
	if result == nil {
		result = goja.Undefined()
	}
	_ = resolve(result)
	return vm.ToValue(p)
}

func errorObject(vm *goja.Runtime, code, operation string, err error) *goja.Object {
	errObj := vm.NewObject()
	_ = errObj.Set("message", err.Error())
	if strings.TrimSpace(code) != "" {
		_ = errObj.Set("code", code)
	}
	if strings.TrimSpace(operation) != "" {
		_ = errObj.Set("operation", operation)
	}
	return errObj
}

func resultToValue(vm *goja.Runtime, result cozoapi.CozoResult) goja.Value {
	obj := vm.NewObject()
	_ = obj.Set("headers", result.Headers)
	_ = obj.Set("rows", result.Rows)
	if result.Took != nil {
		_ = obj.Set("took", *result.Took)
	}
	if result.Count != nil {
		_ = obj.Set("count", *result.Count)
	}
	if result.Next != nil {
		_ = obj.Set("next", result.Next)
	}
	_ = obj.Set("objects", func() []map[string]cozoapi.CozoValue {
		return result.Objects()
	})
	_ = obj.Set("firstObject", func() map[string]cozoapi.CozoValue {
		return result.FirstObject()
	})
	_ = obj.Set("scalar", func() cozoapi.CozoValue {
		return result.Scalar()
	})
	return obj
}

func preparedQueryToValue(vm *goja.Runtime, query cozoapi.PreparedQuery) goja.Value {
	obj := vm.NewObject()
	_ = obj.Set("script", query.Script)
	_ = obj.Set("params", query.Params)
	if query.Opts != nil {
		_ = obj.Set("opts", query.Opts)
	}
	return obj
}

func decodeOpenOptions(vm *goja.Runtime, v goja.Value) (OpenOptions, error) {
	if isAbsent(v) {
		return OpenOptions{Backend: "fake", Options: map[string]any{}}, nil
	}
	obj := v.ToObject(vm)
	opts := OpenOptions{Backend: "fake", Options: map[string]any{}}
	if backend := obj.Get("backend"); !isAbsent(backend) {
		opts.Backend = strings.TrimSpace(backend.String())
	}
	if engine := obj.Get("engine"); !isAbsent(engine) {
		opts.Engine = strings.TrimSpace(engine.String())
	}
	if path := obj.Get("path"); !isAbsent(path) {
		opts.Path = path.String()
	}
	if options := obj.Get("options"); !isAbsent(options) {
		optionsMap := map[string]any{}
		if err := vm.ExportTo(options, &optionsMap); err != nil {
			return OpenOptions{}, fmt.Errorf("decode open options.options: %w", err)
		}
		opts.Options = optionsMap
	}
	if strings.TrimSpace(opts.Backend) == "" {
		opts.Backend = "fake"
	}
	return opts, nil
}

func decodeExecQuery(vm *goja.Runtime, args []goja.Value) (cozoapi.PreparedQuery, error) {
	if len(args) == 0 {
		return cozoapi.PreparedQuery{}, fmt.Errorf("exec requires input")
	}

	// exec(preparedQuery)
	if object, ok := args[0].(*goja.Object); ok && hasOwnProperty(object, "script") {
		return decodePreparedQueryValue(vm, args[0])
	}

	// exec(script, params?, opts?)
	var script string
	if err := vm.ExportTo(args[0], &script); err != nil {
		return cozoapi.PreparedQuery{}, fmt.Errorf("decode exec script: %w", err)
	}
	params := cozoapi.CozoParams{}
	if len(args) > 1 && !goja.IsUndefined(args[1]) && !goja.IsNull(args[1]) {
		if err := vm.ExportTo(args[1], &params); err != nil {
			return cozoapi.PreparedQuery{}, fmt.Errorf("decode exec params: %w", err)
		}
	}
	var opts *cozoapi.QueryOptions
	if len(args) > 2 {
		decoded, err := decodeQueryOptions(vm, args[2])
		if err != nil {
			return cozoapi.PreparedQuery{}, err
		}
		opts = decoded
	}

	return cozoapi.PreparedQuery{
		Script: script,
		Params: params,
		Opts:   opts,
		Meta: cozoapi.QueryMeta{
			UsesSystemOp: strings.Contains(script, "::"),
		},
	}, nil
}

func decodeAtomicInput(vm *goja.Runtime, args []goja.Value) ([]cozoapi.PreparedQuery, *cozoapi.AtomicOptions, error) {
	if len(args) == 0 {
		return nil, nil, fmt.Errorf("atomic requires query array")
	}
	arrayObj := args[0].ToObject(vm)
	length := int(arrayObj.Get("length").ToInteger())
	queries := make([]cozoapi.PreparedQuery, 0, length)
	for i := 0; i < length; i++ {
		item := arrayObj.Get(strconv.Itoa(i))
		query, err := decodePreparedQueryValue(vm, item)
		if err != nil {
			return nil, nil, fmt.Errorf("decode atomic query %d: %w", i, err)
		}
		queries = append(queries, query)
	}

	var opts *cozoapi.AtomicOptions
	if len(args) > 1 && !isAbsent(args[1]) {
		obj := args[1].ToObject(vm)
		decoded := cozoapi.AtomicOptions{}
		if write := obj.Get("write"); !isAbsent(write) {
			w := write.ToBoolean()
			decoded.Write = &w
		}
		opts = &decoded
	}
	return queries, opts, nil
}

func decodePreparedQueryValue(vm *goja.Runtime, v goja.Value) (cozoapi.PreparedQuery, error) {
	if isAbsent(v) {
		return cozoapi.PreparedQuery{}, fmt.Errorf("prepared query cannot be empty")
	}
	obj := v.ToObject(vm)
	scriptValue := obj.Get("script")
	if isAbsent(scriptValue) {
		scriptValue = obj.Get("Script")
	}
	if isAbsent(scriptValue) {
		return cozoapi.PreparedQuery{}, fmt.Errorf("prepared query missing script")
	}

	script := scriptValue.String()
	params := cozoapi.CozoParams{}
	paramsValue := obj.Get("params")
	if isAbsent(paramsValue) {
		paramsValue = obj.Get("Params")
	}
	if !isAbsent(paramsValue) {
		if err := vm.ExportTo(paramsValue, &params); err != nil {
			return cozoapi.PreparedQuery{}, fmt.Errorf("decode prepared query params: %w", err)
		}
	}

	optsValue := obj.Get("opts")
	if isAbsent(optsValue) {
		optsValue = obj.Get("Opts")
	}
	opts, err := decodeQueryOptions(vm, optsValue)
	if err != nil {
		return cozoapi.PreparedQuery{}, fmt.Errorf("decode prepared query options: %w", err)
	}

	query := cozoapi.PreparedQuery{
		Script: script,
		Params: params,
		Opts:   opts,
	}
	if metaValue := obj.Get("meta"); !isAbsent(metaValue) {
		metaObj := metaValue.ToObject(vm)
		if relationsValue := metaObj.Get("relations"); !isAbsent(relationsValue) {
			var relations []string
			if err := vm.ExportTo(relationsValue, &relations); err == nil {
				query.Meta.Relations = relations
			}
		}
		if usesSystemOpValue := metaObj.Get("usesSystemOp"); !isAbsent(usesSystemOpValue) {
			query.Meta.UsesSystemOp = usesSystemOpValue.ToBoolean()
		}
	}

	return query, nil
}

func decodeQueryOptions(vm *goja.Runtime, v goja.Value) (*cozoapi.QueryOptions, error) {
	if isAbsent(v) {
		return nil, nil
	}
	obj := v.ToObject(vm)
	opts := cozoapi.QueryOptions{}
	if limit := obj.Get("limit"); !isAbsent(limit) {
		n := int64(limit.ToInteger())
		opts.Limit = &n
	}
	if offset := obj.Get("offset"); !isAbsent(offset) {
		n := int64(offset.ToInteger())
		opts.Offset = &n
	}
	if timeout := obj.Get("timeoutSec"); !isAbsent(timeout) {
		f := timeout.ToFloat()
		opts.TimeoutSec = &f
	}
	if immutable := obj.Get("immutable"); !isAbsent(immutable) {
		b := immutable.ToBoolean()
		opts.Immutable = &b
	}
	return &opts, nil
}

func decodeTemplateParts(vm *goja.Runtime, v goja.Value) ([]string, bool, error) {
	if isAbsent(v) {
		return nil, false, nil
	}
	obj := v.ToObject(vm)
	lengthValue := obj.Get("length")
	if isAbsent(lengthValue) {
		return nil, false, nil
	}

	length := int(lengthValue.ToInteger())
	if length <= 0 {
		return nil, false, fmt.Errorf("template strings array cannot be empty")
	}
	parts := make([]string, 0, length)
	for i := 0; i < length; i++ {
		parts = append(parts, obj.Get(strconv.Itoa(i)).String())
	}
	return parts, true, nil
}

func decodeMutationInput(vm *goja.Runtime, args []goja.Value) ([]map[string]cozoapi.CozoValue, cozoapi.RelationMutationOptions, error) {
	if len(args) == 0 {
		return nil, cozoapi.RelationMutationOptions{}, fmt.Errorf("relation mutation requires rows")
	}
	rows, err := decodeRows(vm, args[0])
	if err != nil {
		return nil, cozoapi.RelationMutationOptions{}, err
	}

	opts, err := decodeRelationMutationOptions(vm, argAt(args, 1))
	if err != nil {
		return nil, cozoapi.RelationMutationOptions{}, err
	}

	return rows, opts, nil
}

func decodeRows(vm *goja.Runtime, v goja.Value) ([]map[string]cozoapi.CozoValue, error) {
	if rows, matched, err := decodeTupleRowsPayload(vm, v); matched || err != nil {
		return rows, err
	}
	if rawRows, ok := v.Export().([]any); ok && len(rawRows) > 0 {
		if _, tupleLike := rawRows[0].([]any); tupleLike {
			return nil, fmt.Errorf("tuple row arrays require explicit mapping object {headers, rows}")
		}
	}

	var mappedRows []map[string]cozoapi.CozoValue
	if err := vm.ExportTo(v, &mappedRows); err == nil && len(mappedRows) > 0 {
		if looksLikeTupleMappedRows(mappedRows) {
			return nil, fmt.Errorf("tuple row arrays require explicit mapping object {headers, rows}")
		}
		return mappedRows, nil
	}

	var typedMappedRows []map[string]any
	if err := vm.ExportTo(v, &typedMappedRows); err == nil && len(typedMappedRows) > 0 {
		rows := make([]map[string]cozoapi.CozoValue, 0, len(typedMappedRows))
		for _, raw := range typedMappedRows {
			row := map[string]cozoapi.CozoValue{}
			for key, value := range raw {
				row[key] = value
			}
			rows = append(rows, row)
		}
		return rows, nil
	}

	var arrayRows [][]any
	if err := vm.ExportTo(v, &arrayRows); err == nil && len(arrayRows) > 0 {
		return nil, fmt.Errorf("tuple row arrays require explicit mapping object {headers, rows}")
	}

	exported := v.Export()
	rawRows, ok := exported.([]any)
	if !ok {
		return nil, fmt.Errorf("rows must be an array")
	}

	rows := make([]map[string]cozoapi.CozoValue, 0, len(rawRows))
	for i, raw := range rawRows {
		switch typed := raw.(type) {
		case map[string]any:
			row := map[string]cozoapi.CozoValue{}
			for key, value := range typed {
				row[key] = value
			}
			rows = append(rows, row)
		case []any:
			return nil, fmt.Errorf("tuple row arrays require explicit mapping object {headers, rows}")
		default:
			obj := vm.ToValue(raw)
			mapRow := map[string]cozoapi.CozoValue{}
			if err := vm.ExportTo(obj, &mapRow); err != nil {
				return nil, fmt.Errorf("row %d: unsupported row type %T", i, raw)
			}
			rows = append(rows, mapRow)
		}
	}
	return rows, nil
}

func decodeRelationSpec(vm *goja.Runtime, v goja.Value) (cozoapi.RelationSpec, error) {
	if isAbsent(v) {
		return cozoapi.RelationSpec{}, fmt.Errorf("relation spec cannot be empty")
	}
	obj := v.ToObject(vm)

	keysValue := firstPresentField(obj, "keys", "Keys")
	if isAbsent(keysValue) {
		return cozoapi.RelationSpec{}, fmt.Errorf("relation spec requires keys")
	}
	keys, err := decodeStringMap(vm, keysValue, "keys")
	if err != nil {
		return cozoapi.RelationSpec{}, err
	}
	if len(keys) == 0 {
		return cozoapi.RelationSpec{}, fmt.Errorf("relation spec requires non-empty keys")
	}

	values := map[string]string{}
	valuesValue := firstPresentField(obj, "values", "Values")
	if !isAbsent(valuesValue) {
		decodedValues, err := decodeStringMap(vm, valuesValue, "values")
		if err != nil {
			return cozoapi.RelationSpec{}, err
		}
		values = decodedValues
	}

	return cozoapi.RelationSpec{
		Keys:   keys,
		Values: values,
	}, nil
}

func decodeRelationCreateOptions(vm *goja.Runtime, v goja.Value) (cozoapi.RelationCreateOptions, error) {
	opts := cozoapi.RelationCreateOptions{}
	if isAbsent(v) {
		return opts, nil
	}
	obj := v.ToObject(vm)
	replaceValue := firstPresentField(obj, "replace", "Replace")
	if !isAbsent(replaceValue) {
		opts.Replace = replaceValue.ToBoolean()
	}
	return opts, nil
}

func decodeRelationMutationOptions(vm *goja.Runtime, v goja.Value) (cozoapi.RelationMutationOptions, error) {
	opts := cozoapi.RelationMutationOptions{}
	if isAbsent(v) {
		return opts, nil
	}
	obj := v.ToObject(vm)
	returningValue := firstPresentField(obj, "returning", "Returning")
	if !isAbsent(returningValue) {
		opts.Returning = returningValue.ToBoolean()
	}
	return opts, nil
}

func decodeTupleRowsPayload(vm *goja.Runtime, v goja.Value) ([]map[string]cozoapi.CozoValue, bool, error) {
	if isAbsent(v) {
		return nil, false, nil
	}
	obj := v.ToObject(vm)
	headersValue := firstPresentField(obj, "headers", "Headers")
	rowsValue := firstPresentField(obj, "rows", "Rows")
	if isAbsent(headersValue) && isAbsent(rowsValue) {
		return nil, false, nil
	}
	if isAbsent(headersValue) || isAbsent(rowsValue) {
		return nil, true, fmt.Errorf("tuple row payload requires both headers and rows")
	}

	var headers []string
	if err := vm.ExportTo(headersValue, &headers); err != nil {
		return nil, true, fmt.Errorf("decode tuple headers: %w", err)
	}
	if len(headers) == 0 {
		return nil, true, fmt.Errorf("tuple row payload requires non-empty headers")
	}

	normalizedHeaders := make([]string, len(headers))
	seen := map[string]struct{}{}
	for i, header := range headers {
		header = strings.TrimSpace(header)
		if header == "" {
			return nil, true, fmt.Errorf("tuple header at index %d is empty", i)
		}
		if _, ok := seen[header]; ok {
			return nil, true, fmt.Errorf("duplicate tuple header %q", header)
		}
		seen[header] = struct{}{}
		normalizedHeaders[i] = header
	}

	var tupleRows [][]any
	if err := vm.ExportTo(rowsValue, &tupleRows); err != nil {
		return nil, true, fmt.Errorf("decode tuple rows: %w", err)
	}

	out := make([]map[string]cozoapi.CozoValue, 0, len(tupleRows))
	for i, tuple := range tupleRows {
		if len(tuple) != len(normalizedHeaders) {
			return nil, true, fmt.Errorf(
				"tuple row %d length %d does not match headers length %d",
				i, len(tuple), len(normalizedHeaders),
			)
		}
		row := make(map[string]cozoapi.CozoValue, len(normalizedHeaders))
		for j, header := range normalizedHeaders {
			row[header] = tuple[j]
		}
		out = append(out, row)
	}
	return out, true, nil
}

func decodeStringMap(vm *goja.Runtime, v goja.Value, fieldName string) (map[string]string, error) {
	out := map[string]string{}
	if err := vm.ExportTo(v, &out); err == nil {
		return out, nil
	}

	raw := map[string]any{}
	if err := vm.ExportTo(v, &raw); err != nil {
		return nil, fmt.Errorf("decode relation spec %s: %w", fieldName, err)
	}
	for key, value := range raw {
		if value == nil {
			out[key] = ""
			continue
		}
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("relation spec %s.%s must be a string", fieldName, key)
		}
		out[key] = s
	}
	return out, nil
}

func firstPresentField(obj *goja.Object, names ...string) goja.Value {
	for _, name := range names {
		v := obj.Get(name)
		if !isAbsent(v) {
			return v
		}
	}
	return nil
}

func argAt(args []goja.Value, idx int) goja.Value {
	if idx < 0 || idx >= len(args) {
		return nil
	}
	return args[idx]
}

func decodeStringArray(vm *goja.Runtime, args []goja.Value) ([]string, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("expected string array argument")
	}
	var items []string
	if err := vm.ExportTo(args[0], &items); err != nil {
		return nil, fmt.Errorf("decode string array: %w", err)
	}
	return items, nil
}

func looksLikeTupleMappedRows(rows []map[string]cozoapi.CozoValue) bool {
	for _, row := range rows {
		if len(row) == 0 {
			return false
		}
		for key := range row {
			if _, err := strconv.Atoi(key); err != nil {
				return false
			}
		}
	}
	return true
}

func hasOwnProperty(obj *goja.Object, key string) bool {
	v := obj.Get(key)
	return !isAbsent(v)
}

func isAbsent(v goja.Value) bool {
	return v == nil || goja.IsUndefined(v) || goja.IsNull(v)
}
