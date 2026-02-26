package cozoapi

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var relationNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type RelationSpec struct {
	Keys   map[string]string
	Values map[string]string
}

type RelationCreateOptions struct {
	Replace bool
}

type RelationMutationOptions struct {
	Returning bool
}

type AccessLevel string

const (
	AccessNormal    AccessLevel = "normal"
	AccessProtected AccessLevel = "protected"
	AccessReadOnly  AccessLevel = "read_only"
	AccessHidden    AccessLevel = "hidden"
)

func (l AccessLevel) IsValid() bool {
	switch l {
	case AccessNormal, AccessProtected, AccessReadOnly, AccessHidden:
		return true
	default:
		return false
	}
}

func CompileRelationCreate(name string, spec RelationSpec, opts RelationCreateOptions) (PreparedQuery, error) {
	if err := validateRelationName(name); err != nil {
		return PreparedQuery{}, err
	}
	if len(spec.Keys) == 0 {
		return PreparedQuery{}, fmt.Errorf("relation create requires at least one key column")
	}

	keyCols := make([]string, 0, len(spec.Keys))
	for _, col := range sortedKeys(spec.Keys) {
		typ := strings.TrimSpace(spec.Keys[col])
		if typ == "" {
			keyCols = append(keyCols, col)
		} else {
			keyCols = append(keyCols, fmt.Sprintf("%s: %s", col, typ))
		}
	}

	valueCols := make([]string, 0, len(spec.Values))
	for _, col := range sortedKeys(spec.Values) {
		typ := strings.TrimSpace(spec.Values[col])
		if typ == "" {
			valueCols = append(valueCols, col)
		} else {
			valueCols = append(valueCols, fmt.Sprintf("%s: %s", col, typ))
		}
	}

	mode := "create"
	if opts.Replace {
		mode = "replace"
	}

	specBody := strings.Join(keyCols, ", ")
	if len(valueCols) > 0 {
		specBody = specBody + " => " + strings.Join(valueCols, ", ")
	}

	return PreparedQuery{
		Script: fmt.Sprintf(":%s %s {%s}", mode, name, specBody),
		Params: CozoParams{},
		Meta: QueryMeta{
			Relations: []string{name},
		},
	}, nil
}

func CompileRelationPut(name string, rows []map[string]CozoValue, opts RelationMutationOptions) (PreparedQuery, error) {
	return compileRelationMutation(name, "put", rows, opts)
}

func CompileRelationInsert(name string, rows []map[string]CozoValue, opts RelationMutationOptions) (PreparedQuery, error) {
	return compileRelationMutation(name, "insert", rows, opts)
}

func CompileRelationUpdate(name string, rows []map[string]CozoValue, opts RelationMutationOptions) (PreparedQuery, error) {
	return compileRelationMutation(name, "update", rows, opts)
}

func CompileRelationRm(name string, keys []map[string]CozoValue, opts RelationMutationOptions) (PreparedQuery, error) {
	return compileRelationMutation(name, "rm", keys, opts)
}

func CompileRelationDel(name string, rows []map[string]CozoValue, opts RelationMutationOptions) (PreparedQuery, error) {
	return compileRelationMutation(name, "delete", rows, opts)
}

func CompileRelationGet(name string, key map[string]CozoValue) (PreparedQuery, error) {
	if err := validateRelationName(name); err != nil {
		return PreparedQuery{}, err
	}
	if len(key) == 0 {
		return PreparedQuery{}, fmt.Errorf("relation get requires a non-empty key")
	}
	keyCols := sortedKeys(key)
	if len(keyCols) == 0 {
		return PreparedQuery{}, fmt.Errorf("relation get requires at least one key column")
	}

	return compileRelationGetWithColumns(name, keyCols, key)
}

func compileRelationGetWithColumns(name string, columns []string, key map[string]CozoValue) (PreparedQuery, error) {
	if err := validateRelationName(name); err != nil {
		return PreparedQuery{}, err
	}
	if len(columns) == 0 {
		return PreparedQuery{}, fmt.Errorf("relation get requires at least one projection column")
	}
	if len(key) == 0 {
		return PreparedQuery{}, fmt.Errorf("relation get requires a non-empty key")
	}

	normalizedCols := make([]string, 0, len(columns))
	seen := map[string]struct{}{}
	for _, col := range columns {
		col = strings.TrimSpace(col)
		if col == "" {
			continue
		}
		if _, ok := seen[col]; ok {
			continue
		}
		seen[col] = struct{}{}
		normalizedCols = append(normalizedCols, col)
	}
	if len(normalizedCols) == 0 {
		return PreparedQuery{}, fmt.Errorf("relation get requires non-empty projection columns")
	}

	clauses := make([]string, 0, len(key))
	params := make(CozoParams, len(key))
	for _, col := range sortedKeys(key) {
		paramName := "k_" + col
		clauses = append(clauses, fmt.Sprintf("%s == $%s", col, paramName))
		params[paramName] = key[col]
	}

	return PreparedQuery{
		Script: fmt.Sprintf(
			"?[%s] := *%s{%s}, %s",
			strings.Join(normalizedCols, ", "),
			name,
			strings.Join(normalizedCols, ", "),
			strings.Join(clauses, ", "),
		),
		Params: params,
		Meta: QueryMeta{
			Relations: []string{name},
		},
	}, nil
}

func relationColumnNames(result CozoResult) ([]string, error) {
	columnHeaderIdx := -1
	for idx, header := range result.Headers {
		if strings.EqualFold(strings.TrimSpace(header), "column") {
			columnHeaderIdx = idx
			break
		}
	}
	if columnHeaderIdx < 0 {
		return nil, fmt.Errorf("relation columns response missing 'column' header")
	}

	out := make([]string, 0, len(result.Rows))
	for i, row := range result.Rows {
		if columnHeaderIdx >= len(row) {
			return nil, fmt.Errorf("relation columns row %d missing column name", i)
		}
		name, ok := row[columnHeaderIdx].(string)
		if !ok {
			return nil, fmt.Errorf("relation columns row %d has non-string column name", i)
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, fmt.Errorf("relation columns row %d has empty column name", i)
		}
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("relation has no columns")
	}
	return out, nil
}

func CompileRelationColumns(name string) (PreparedQuery, error) {
	if err := validateRelationName(name); err != nil {
		return PreparedQuery{}, err
	}
	return PreparedQuery{
		Script: fmt.Sprintf("::columns %s", name),
		Params: CozoParams{},
		Meta: QueryMeta{
			Relations:    []string{name},
			UsesSystemOp: true,
		},
	}, nil
}

func CompileRelationIndices(name string) (PreparedQuery, error) {
	if err := validateRelationName(name); err != nil {
		return PreparedQuery{}, err
	}
	return PreparedQuery{
		Script: fmt.Sprintf("::indices %s", name),
		Params: CozoParams{},
		Meta: QueryMeta{
			Relations:    []string{name},
			UsesSystemOp: true,
		},
	}, nil
}

func CompileRelationAccess(name string, level AccessLevel) (PreparedQuery, error) {
	if err := validateRelationName(name); err != nil {
		return PreparedQuery{}, err
	}
	if !level.IsValid() {
		return PreparedQuery{}, fmt.Errorf("invalid relation access level %q", level)
	}
	return PreparedQuery{
		Script: fmt.Sprintf("::access_level %s %s", name, string(level)),
		Params: CozoParams{},
		Meta: QueryMeta{
			Relations:    []string{name},
			UsesSystemOp: true,
		},
	}, nil
}

func compileRelationMutation(name, op string, rows []map[string]CozoValue, opts RelationMutationOptions) (PreparedQuery, error) {
	if err := validateRelationName(name); err != nil {
		return PreparedQuery{}, err
	}
	if len(rows) == 0 {
		return PreparedQuery{}, fmt.Errorf("relation mutation %s requires at least one row", op)
	}

	columnSet := map[string]struct{}{}
	for _, row := range rows {
		for col := range row {
			columnSet[col] = struct{}{}
		}
	}
	if len(columnSet) == 0 {
		return PreparedQuery{}, fmt.Errorf("relation mutation %s requires at least one column", op)
	}
	columns := make([]string, 0, len(columnSet))
	for col := range columnSet {
		columns = append(columns, col)
	}
	slices.Sort(columns)

	params := make(CozoParams, len(rows)*len(columns))
	rowStrings := make([]string, 0, len(rows))
	for rowIdx, row := range rows {
		cells := make([]string, 0, len(columns))
		for _, col := range columns {
			paramName := fmt.Sprintf("r%d_%s", rowIdx+1, col)
			cells = append(cells, "$"+paramName)
			params[paramName] = row[col]
		}
		rowStrings = append(rowStrings, "["+strings.Join(cells, ", ")+"]")
	}

	lines := []string{
		fmt.Sprintf("?[%s] <- [%s]", strings.Join(columns, ", "), strings.Join(rowStrings, ", ")),
		fmt.Sprintf(":%s %s {%s}", op, name, strings.Join(columns, ", ")),
	}
	if opts.Returning {
		lines = append(lines, ":returning")
	}

	return PreparedQuery{
		Script: joinLines(lines),
		Params: params,
		Meta: QueryMeta{
			Relations: []string{name},
		},
	}, nil
}

func validateRelationName(name string) error {
	if !relationNameRe.MatchString(name) {
		return fmt.Errorf("invalid relation name %q", name)
	}
	return nil
}

type RelationHandle struct {
	db   *DB
	name string
}

func (r *RelationHandle) Name() string {
	if r == nil {
		return ""
	}
	return r.name
}

func (r *RelationHandle) Create(ctx context.Context, spec RelationSpec, opts RelationCreateOptions) error {
	query, err := CompileRelationCreate(r.name, spec, opts)
	if err != nil {
		return err
	}
	_, err = r.db.ExecPrepared(ctx, query)
	return err
}

func (r *RelationHandle) Put(ctx context.Context, rows []map[string]CozoValue, opts RelationMutationOptions) (CozoResult, error) {
	query, err := CompileRelationPut(r.name, rows, opts)
	if err != nil {
		return CozoResult{}, err
	}
	return r.db.ExecPrepared(ctx, query)
}

func (r *RelationHandle) Insert(ctx context.Context, rows []map[string]CozoValue, opts RelationMutationOptions) (CozoResult, error) {
	query, err := CompileRelationInsert(r.name, rows, opts)
	if err != nil {
		return CozoResult{}, err
	}
	return r.db.ExecPrepared(ctx, query)
}

func (r *RelationHandle) Update(ctx context.Context, rows []map[string]CozoValue, opts RelationMutationOptions) (CozoResult, error) {
	query, err := CompileRelationUpdate(r.name, rows, opts)
	if err != nil {
		return CozoResult{}, err
	}
	return r.db.ExecPrepared(ctx, query)
}

func (r *RelationHandle) Rm(ctx context.Context, keys []map[string]CozoValue, opts RelationMutationOptions) (CozoResult, error) {
	query, err := CompileRelationRm(r.name, keys, opts)
	if err != nil {
		return CozoResult{}, err
	}
	return r.db.ExecPrepared(ctx, query)
}

func (r *RelationHandle) Del(ctx context.Context, rows []map[string]CozoValue, opts RelationMutationOptions) (CozoResult, error) {
	query, err := CompileRelationDel(r.name, rows, opts)
	if err != nil {
		return CozoResult{}, err
	}
	return r.db.ExecPrepared(ctx, query)
}

func (r *RelationHandle) Get(ctx context.Context, key map[string]CozoValue) (map[string]CozoValue, error) {
	query, err := CompileRelationGet(r.name, key)
	if err != nil {
		return nil, err
	}

	// Best effort: query relation columns first so Get can return a full row on
	// real backends. Keep fake backend behavior unchanged for deterministic tests.
	// If this path fails (for example due policy/system-op restrictions), fallback
	// to key-column projection query generated by CompileRelationGet.
	if r != nil && r.db != nil && r.db.backend != nil && !strings.EqualFold(r.db.backend.Name(), "fake") {
		if columnsResult, err := r.Columns(ctx); err == nil {
			if names, parseErr := relationColumnNames(columnsResult); parseErr == nil {
				if richerQuery, qErr := compileRelationGetWithColumns(r.name, names, key); qErr == nil {
					query = richerQuery
				}
			}
		}
	}

	result, err := r.db.ExecPrepared(ctx, query)
	if err != nil {
		return nil, err
	}
	return result.FirstObject(), nil
}

func (r *RelationHandle) Columns(ctx context.Context) (CozoResult, error) {
	query, err := CompileRelationColumns(r.name)
	if err != nil {
		return CozoResult{}, err
	}
	return r.db.ExecPrepared(ctx, query)
}

func (r *RelationHandle) Indices(ctx context.Context) (CozoResult, error) {
	query, err := CompileRelationIndices(r.name)
	if err != nil {
		return CozoResult{}, err
	}
	return r.db.ExecPrepared(ctx, query)
}

func (r *RelationHandle) Access(ctx context.Context, level AccessLevel) error {
	query, err := CompileRelationAccess(r.name, level)
	if err != nil {
		return err
	}
	_, err = r.db.ExecPrepared(ctx, query)
	return err
}
