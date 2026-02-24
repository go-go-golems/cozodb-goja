package cozoapi

import (
	"fmt"
	"strings"
)

func CompileTemplate(parts []string, values []CozoValue, opts *QueryOptions) (PreparedQuery, error) {
	if len(parts) == 0 {
		return PreparedQuery{}, fmt.Errorf("template parts cannot be empty")
	}
	if len(parts) != len(values)+1 {
		return PreparedQuery{}, fmt.Errorf("template arity mismatch: parts=%d values=%d", len(parts), len(values))
	}

	params := make(CozoParams, len(values))
	var b strings.Builder
	for i := range values {
		b.WriteString(parts[i])
		name := fmt.Sprintf("p%d", i+1)
		params[name] = values[i]
		b.WriteString("$")
		b.WriteString(name)
	}
	b.WriteString(parts[len(parts)-1])

	query := PreparedQuery{
		Script: strings.TrimSpace(b.String()),
		Params: params,
		Meta: QueryMeta{
			UsesSystemOp: strings.Contains(b.String(), "::"),
		},
	}
	if opts != nil {
		cloned := opts.Clone()
		query.Opts = &cloned
	}
	return query, nil
}
