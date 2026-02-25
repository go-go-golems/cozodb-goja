//go:build cozo_cgo

package cozocgo

import (
	"context"
	"fmt"
	"strings"

	cozo "github.com/cozodb/cozo-lib-go"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

type backend struct {
	db cozo.CozoDB
}

var _ cozoapi.Backend = (*backend)(nil)

func Open(_ context.Context, opts OpenOptions) (cozoapi.Backend, error) {
	engine := strings.TrimSpace(opts.Engine)
	if engine == "" {
		engine = "mem"
	}

	path := strings.TrimSpace(opts.Path)
	options := cozo.Map{}
	for key, value := range opts.Options {
		options[key] = value
	}
	if len(options) == 0 {
		options = nil
	}

	db, err := cozo.New(engine, path, options)
	if err != nil {
		return nil, fmt.Errorf("open cozo db: %w", err)
	}
	return &backend{db: db}, nil
}

func (b *backend) Name() string {
	return "cozocgo"
}

func (b *backend) Capabilities() cozoapi.BackendCapabilities {
	return cozoapi.BackendCapabilities{
		Persistence:   true,
		BackupRestore: true,
		Callbacks:     false,
		NamedRules:    false,
	}
}

func (b *backend) Exec(ctx context.Context, script string, params cozoapi.CozoParams, opts cozoapi.QueryOptions) (cozoapi.CozoResult, error) {
	if err := ctx.Err(); err != nil {
		return cozoapi.CozoResult{}, err
	}

	query := strings.TrimSpace(script)
	if directives := opts.CozoDirectives(); directives != "" {
		query = query + "\n" + directives
	}

	cozoParams := cozo.Map{}
	for key, value := range params {
		cozoParams[key] = value
	}
	if len(cozoParams) == 0 {
		cozoParams = nil
	}

	res, err := b.db.Run(query, cozoParams)
	if err != nil {
		return cozoapi.CozoResult{}, fmt.Errorf("run query: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return cozoapi.CozoResult{}, err
	}

	rows := make([]cozoapi.CozoRow, 0, len(res.Rows))
	for _, row := range res.Rows {
		converted := make(cozoapi.CozoRow, len(row))
		for i, v := range row {
			converted[i] = v
		}
		rows = append(rows, converted)
	}

	took := res.Took
	return cozoapi.CozoResult{
		Headers: res.Headers,
		Rows:    rows,
		Took:    &took,
	}, nil
}

func (b *backend) Export(ctx context.Context, relations []string) (map[string]cozoapi.RelationRows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	raw, err := b.db.ExportRelations(relations)
	if err != nil {
		return nil, fmt.Errorf("export relations: %w", err)
	}
	out := map[string]cozoapi.RelationRows{}
	for relation, payload := range raw {
		payloadMap, ok := payload.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("export payload for %q has unsupported type %T", relation, payload)
		}

		headersAny, ok := payloadMap["headers"].([]any)
		if !ok {
			return nil, fmt.Errorf("export payload for %q missing headers", relation)
		}
		headers := make([]string, 0, len(headersAny))
		for _, header := range headersAny {
			headers = append(headers, fmt.Sprintf("%v", header))
		}

		rowsAny, ok := payloadMap["rows"].([]any)
		if !ok {
			return nil, fmt.Errorf("export payload for %q missing rows", relation)
		}
		rows := make([]cozoapi.CozoRow, 0, len(rowsAny))
		for _, rowAny := range rowsAny {
			rowSlice, ok := rowAny.([]any)
			if !ok {
				return nil, fmt.Errorf("export row for %q has unsupported type %T", relation, rowAny)
			}
			row := make(cozoapi.CozoRow, len(rowSlice))
			for i, v := range rowSlice {
				row[i] = v
			}
			rows = append(rows, row)
		}

		out[relation] = cozoapi.RelationRows{
			Headers: headers,
			Rows:    rows,
		}
	}
	return out, nil
}

func (b *backend) Import(ctx context.Context, data map[string]cozoapi.RelationRows) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	payload := cozo.Map{}
	for relation, relationRows := range data {
		rows := make([][]any, 0, len(relationRows.Rows))
		for _, row := range relationRows.Rows {
			converted := make([]any, len(row))
			for i, v := range row {
				converted[i] = v
			}
			rows = append(rows, converted)
		}
		payload[relation] = map[string]any{
			"headers": relationRows.Headers,
			"rows":    rows,
		}
	}
	return b.db.ImportRelations(payload)
}

func (b *backend) Close(context.Context) error {
	b.db.Close()
	return nil
}
