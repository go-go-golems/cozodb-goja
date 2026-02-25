//go:build cozo_cgo

package cozocgo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	cozo "github.com/kraklabs/cie/pkg/cozodb"
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
	options := map[string]any{}
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

	cozoParams := map[string]any{}
	for key, value := range params {
		cozoParams[key] = value
	}
	if len(cozoParams) == 0 {
		cozoParams = nil
	}

	var (
		res cozo.NamedRows
		err error
	)
	if opts.Immutable != nil && *opts.Immutable {
		res, err = b.db.RunReadOnly(query, cozoParams)
	} else {
		res, err = b.db.Run(query, cozoParams)
	}
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

	return cozoapi.CozoResult{
		Headers: res.Headers,
		Rows:    rows,
	}, nil
}

func (b *backend) Export(ctx context.Context, relations []string) (map[string]cozoapi.RelationRows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	reqBytes, err := json.Marshal(map[string]any{"relations": relations})
	if err != nil {
		return nil, fmt.Errorf("marshal export request: %w", err)
	}
	rawJSON, err := b.db.ExportRelations(string(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("export relations: %w", err)
	}

	var response struct {
		OK   bool `json:"ok"`
		Data map[string]struct {
			Headers []string `json:"headers"`
			Rows    [][]any  `json:"rows"`
		} `json:"data"`
		Message string `json:"message"`
		Display string `json:"display"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &response); err != nil {
		return nil, fmt.Errorf("parse export response: %w", err)
	}
	if !response.OK {
		msg := strings.TrimSpace(response.Message)
		if msg == "" {
			msg = strings.TrimSpace(response.Display)
		}
		if msg == "" {
			msg = "export failed"
		}
		return nil, fmt.Errorf("export relations: %s", msg)
	}

	out := map[string]cozoapi.RelationRows{}
	for relation, payload := range response.Data {
		rows := make([]cozoapi.CozoRow, 0, len(payload.Rows))
		for _, rowSlice := range payload.Rows {
			row := make(cozoapi.CozoRow, len(rowSlice))
			for i, v := range rowSlice {
				row[i] = v
			}
			rows = append(rows, row)
		}

		out[relation] = cozoapi.RelationRows{
			Headers: payload.Headers,
			Rows:    rows,
		}
	}
	return out, nil
}

func (b *backend) Import(ctx context.Context, data map[string]cozoapi.RelationRows) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	payload := map[string]any{}
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
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal import payload: %w", err)
	}
	return b.db.ImportRelations(string(payloadBytes))
}

func (b *backend) Close(context.Context) error {
	b.db.Close()
	return nil
}
