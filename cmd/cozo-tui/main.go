package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/cozodb-goja/internal/tui/app"
	"github.com/go-go-golems/cozodb-goja/internal/tui/seeddata"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/cozocgo"
)

func main() {
	var (
		dbEngine = flag.String("engine", "mem", "CozoDB engine (mem, sqlite)")
		dbPath   = flag.String("db", "", "Path to CozoDB database file")
		seed     = flag.Bool("seed", false, "Seed sample data if database is empty")
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

	// For mem engine, always seed; for persistent engines, only if --seed
	if *dbEngine == "mem" || *seed {
		if err := seeddata.SeedIfEmpty(ctx, db); err != nil {
			fmt.Fprintf(os.Stderr, "error seeding database: %v\n", err)
			os.Exit(1)
		}
	}

	m := app.New(db)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
