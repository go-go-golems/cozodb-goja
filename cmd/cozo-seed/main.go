package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/go-go-golems/cozodb-goja/internal/tui/seeddata"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/cozocgo"
)

func main() {
	var (
		dbEngine = flag.String("engine", "sqlite", "CozoDB engine (mem, sqlite)")
		dbPath   = flag.String("db", "cozo-explorer.db", "Path to CozoDB database file")
		force    = flag.Bool("force", false, "Re-seed even if data already exists")
	)
	flag.Parse()

	ctx := context.Background()

	backend, err := cozocgo.Open(ctx, cozocgo.OpenOptions{
		Engine: *dbEngine,
		Path:   *dbPath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}

	policy := cozoapi.DefaultPolicy()
	policy.AllowSystemOps = true
	db, err := cozoapi.Open(backend, policy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close(ctx)

	has, err := seeddata.HasData(ctx, db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error checking data: %v\n", err)
		os.Exit(1)
	}

	if has && !*force {
		fmt.Println("Database already has data. Use --force to re-seed.")
		os.Exit(0)
	}

	if !has {
		fmt.Printf("Creating schema in %s (%s)...\n", *dbPath, *dbEngine)
		if err := seeddata.CreateSchema(ctx, db); err != nil {
			fmt.Fprintf(os.Stderr, "error creating schema: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Inserting sample data...")
	if err := seeddata.InsertSampleData(ctx, db); err != nil {
		fmt.Fprintf(os.Stderr, "error inserting data: %v\n", err)
		os.Exit(1)
	}

	// Verify
	res, err := db.ExecScript(ctx, `?[count(id)] := *person{id}`, nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error verifying: %v\n", err)
		os.Exit(1)
	}
	persons := 0
	if len(res.Rows) > 0 && len(res.Rows[0]) > 0 {
		if v, ok := res.Rows[0][0].(float64); ok {
			persons = int(v)
		}
	}

	res, err = db.ExecScript(ctx, `?[count(id)] := *relationship{id}`, nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error verifying: %v\n", err)
		os.Exit(1)
	}
	rels := 0
	if len(res.Rows) > 0 && len(res.Rows[0]) > 0 {
		if v, ok := res.Rows[0][0].(float64); ok {
			rels = int(v)
		}
	}

	fmt.Printf("Done. Seeded %d persons, %d relationships.\n", persons, rels)
	fmt.Printf("Database: %s\n", *dbPath)
	fmt.Printf("\nRun the TUI:\n  cozo-tui --engine %s --db %s\n", *dbEngine, *dbPath)
}
