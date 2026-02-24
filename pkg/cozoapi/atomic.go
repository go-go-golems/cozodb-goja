package cozoapi

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var paramTokenRe = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)

type AtomicOptions struct {
	Write *bool
}

func CompileAtomic(queries []PreparedQuery, _ *AtomicOptions) (PreparedQuery, error) {
	if len(queries) == 0 {
		return PreparedQuery{}, fmt.Errorf("atomic requires at least one query")
	}

	params := CozoParams{}
	relations := make([]string, 0)
	seenRelations := map[string]struct{}{}
	scripts := make([]string, 0, len(queries))
	usesSystemOp := false

	for i, q := range queries {
		if strings.TrimSpace(q.Script) == "" {
			return PreparedQuery{}, fmt.Errorf("atomic query %d has empty script", i)
		}

		prefix := fmt.Sprintf("q%d_", i+1)
		script, rewrite := namespaceScriptParams(q.Script, prefix)

		block := "{\n" + strings.TrimSpace(script) + "\n}"
		scripts = append(scripts, block)

		if q.Params != nil {
			for old, newName := range rewrite {
				value, ok := q.Params[old]
				if !ok {
					continue
				}
				params[newName] = value
			}
			for key, value := range q.Params {
				if _, ok := rewrite[key]; ok {
					continue
				}
				// If params were passed but not referenced by script, keep them namespaced too.
				params[prefix+key] = value
			}
		}

		if q.Meta.UsesSystemOp || strings.Contains(q.Script, "::") {
			usesSystemOp = true
		}
		for _, rel := range q.Meta.Relations {
			if _, ok := seenRelations[rel]; ok {
				continue
			}
			seenRelations[rel] = struct{}{}
			relations = append(relations, rel)
		}
	}
	slices.Sort(relations)

	return PreparedQuery{
		Script: strings.Join(scripts, "\n"),
		Params: params,
		Meta: QueryMeta{
			Relations:    relations,
			UsesSystemOp: usesSystemOp,
		},
	}, nil
}

func namespaceScriptParams(script, prefix string) (string, map[string]string) {
	rewrite := map[string]string{}
	namespaced := paramTokenRe.ReplaceAllStringFunc(script, func(raw string) string {
		name := strings.TrimPrefix(raw, "$")
		if _, ok := rewrite[name]; !ok {
			rewrite[name] = prefix + name
		}
		return "$" + rewrite[name]
	})
	return namespaced, rewrite
}
