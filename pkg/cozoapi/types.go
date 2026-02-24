package cozoapi

import (
	"fmt"
	"sort"
)

type CozoValue = any

type CozoParams map[string]CozoValue

type CozoRow []CozoValue

type CozoHeaders []string

type CozoResult struct {
	Headers CozoHeaders `json:"headers"`
	Rows    []CozoRow   `json:"rows"`
	Took    *float64    `json:"took,omitempty"`
	Count   *int64      `json:"count,omitempty"`
	Next    any         `json:"next,omitempty"`
}

func (r CozoResult) Objects() []map[string]CozoValue {
	out := make([]map[string]CozoValue, 0, len(r.Rows))
	for _, row := range r.Rows {
		obj := make(map[string]CozoValue, len(r.Headers))
		for i, header := range r.Headers {
			if i < len(row) {
				obj[header] = row[i]
				continue
			}
			obj[header] = nil
		}
		out = append(out, obj)
	}
	return out
}

func (r CozoResult) FirstObject() map[string]CozoValue {
	if len(r.Rows) == 0 {
		return nil
	}
	row := r.Rows[0]
	obj := make(map[string]CozoValue, len(r.Headers))
	for i, header := range r.Headers {
		if i < len(row) {
			obj[header] = row[i]
			continue
		}
		obj[header] = nil
	}
	return obj
}

func (r CozoResult) Scalar() CozoValue {
	if len(r.Rows) == 0 || len(r.Rows[0]) == 0 {
		return nil
	}
	return r.Rows[0][0]
}

type QueryOptions struct {
	Limit      *int64   `json:"limit,omitempty"`
	Offset     *int64   `json:"offset,omitempty"`
	TimeoutSec *float64 `json:"timeoutSec,omitempty"`
	Immutable  *bool    `json:"immutable,omitempty"`
}

func (o QueryOptions) Clone() QueryOptions {
	cloned := QueryOptions{}
	if o.Limit != nil {
		v := *o.Limit
		cloned.Limit = &v
	}
	if o.Offset != nil {
		v := *o.Offset
		cloned.Offset = &v
	}
	if o.TimeoutSec != nil {
		v := *o.TimeoutSec
		cloned.TimeoutSec = &v
	}
	if o.Immutable != nil {
		v := *o.Immutable
		cloned.Immutable = &v
	}
	return cloned
}

func (o QueryOptions) IsZero() bool {
	return o.Limit == nil && o.Offset == nil && o.TimeoutSec == nil && o.Immutable == nil
}

func (o QueryOptions) CozoDirectives() string {
	lines := make([]string, 0, 3)
	if o.Limit != nil {
		lines = append(lines, fmt.Sprintf(":limit %d", *o.Limit))
	}
	if o.Offset != nil {
		lines = append(lines, fmt.Sprintf(":offset %d", *o.Offset))
	}
	if o.TimeoutSec != nil {
		lines = append(lines, fmt.Sprintf(":timeout %g", *o.TimeoutSec))
	}
	return joinLines(lines)
}

type PreparedQuery struct {
	Script string
	Params CozoParams
	Opts   *QueryOptions
	Meta   QueryMeta
}

type QueryMeta struct {
	Relations    []string
	UsesSystemOp bool
}

type BackendCapabilities struct {
	Persistence   bool
	BackupRestore bool
	Callbacks     bool
	NamedRules    bool
}

type RelationRows struct {
	Headers []string  `json:"headers"`
	Rows    []CozoRow `json:"rows"`
}

func PtrInt64(v int64) *int64 { return &v }

func PtrFloat64(v float64) *float64 { return &v }

func PtrBool(v bool) *bool { return &v }

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
