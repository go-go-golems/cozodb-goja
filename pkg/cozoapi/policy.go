package cozoapi

import (
	"fmt"
	"strings"
)

type Policy struct {
	QueryOptions   QueryOptionPolicy
	AllowSystemOps bool
	DeniedRelation map[string]struct{}
}

func DefaultPolicy() Policy {
	return Policy{
		AllowSystemOps: true,
		DeniedRelation: map[string]struct{}{},
	}
}

func (p Policy) WithDeniedRelations(relations ...string) Policy {
	if p.DeniedRelation == nil {
		p.DeniedRelation = map[string]struct{}{}
	}
	for _, relation := range relations {
		relation = strings.TrimSpace(relation)
		if relation == "" {
			continue
		}
		p.DeniedRelation[relation] = struct{}{}
	}
	return p
}

func (p Policy) Enforce(query PreparedQuery) (PreparedQuery, error) {
	out := query
	if out.Opts != nil {
		opts := NormalizeQueryOptions(out.Opts, p.QueryOptions)
		out.Opts = &opts
	}

	usesSystemOps := out.Meta.UsesSystemOp || strings.Contains(out.Script, "::")
	if usesSystemOps && !p.AllowSystemOps {
		return PreparedQuery{}, fmt.Errorf("policy rejected system operation")
	}

	if len(p.DeniedRelation) > 0 {
		for _, relation := range out.Meta.Relations {
			if _, denied := p.DeniedRelation[relation]; denied {
				return PreparedQuery{}, fmt.Errorf("policy rejected relation %q", relation)
			}
		}
	}

	return out, nil
}
