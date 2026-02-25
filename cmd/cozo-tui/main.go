package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/cozodb-goja/internal/tui/app"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi/cozocgo"
)

func main() {
	var (
		dbEngine = flag.String("engine", "mem", "CozoDB engine (mem, sqlite, rocksdb)")
		dbPath   = flag.String("db", "", "Path to CozoDB database file")
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

	if err := seedIfEmpty(ctx, db); err != nil {
		fmt.Fprintf(os.Stderr, "error seeding database: %v\n", err)
		os.Exit(1)
	}

	m := app.New(db)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func seedIfEmpty(ctx context.Context, db *cozoapi.DB) error {
	// Check if person relation exists
	res, err := db.ExecScript(ctx, "::relations", nil, nil)
	if err != nil {
		return fmt.Errorf("list relations: %w", err)
	}

	hasPersons := false
	for _, row := range res.Rows {
		if len(row) > 0 {
			if name, ok := row[0].(string); ok && name == "person" {
				hasPersons = true
				break
			}
		}
	}

	if hasPersons {
		return nil
	}

	// Create schema
	schema := []string{
		`:create person {id: String => name: String, description: String, first_mentioned: String}`,
		`:create relationship {id: String, from_person: String, to_person: String, timestamp: String => relationship_type: String, description: String, sentiment: String, strength: Float}`,
		`:create behavior {id: String, person_id: String, timestamp: String => behavior_type: String, description: String}`,
		`:create event {id: String, timestamp: String => description: String}`,
	}
	for _, s := range schema {
		if _, err := db.ExecScript(ctx, s, nil, nil); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
	}

	// Seed sample data
	seeds := []string{
		`?[id, name, description, first_mentioned] <- [
			["sarah_martinez", "Sarah Martinez", "Senior scientist and team lead", "2023-01"],
			["james_liu", "James Liu", "Lab director and mentor", "2023-01"],
			["maya_patel", "Maya Patel", "Graduate student in AI research", "2023-03"],
			["alex_kim", "Alex Kim", "Postdoctoral researcher", "2023-06"],
			["tom_rivera", "Tom Rivera", "Lab technician", "2023-01"]
		]
		:put person {id => name, description, first_mentioned}`,

		`?[id, from_person, to_person, timestamp, relationship_type, description, sentiment, strength] <- [
			["r1", "sarah_martinez", "james_liu", "2023-01", "mentor", "Chen mentors Liu in research methodology", "positive", 0.9],
			["r1", "sarah_martinez", "james_liu", "2023-06", "mentor", "Continued mentoring during sabbatical", "positive", 0.85],
			["r1", "sarah_martinez", "james_liu", "2024-01", "mentor", "Strengthened after joint paper", "positive", 0.92],
			["r2", "sarah_martinez", "maya_patel", "2023-03", "collaborator", "Research collaboration on AI project", "positive", 0.8],
			["r3", "sarah_martinez", "alex_kim", "2023-06", "mentor", "Mentoring new postdoc", "positive", 0.75],
			["r4", "james_liu", "maya_patel", "2023-03", "supervisor", "PhD supervision", "neutral", 0.7],
			["r5", "maya_patel", "tom_rivera", "2023-09", "collaborator", "Lab equipment collaboration", "positive", 0.6],
			["r6", "alex_kim", "tom_rivera", "2023-09", "peer", "Peer researchers", "positive", 0.5],
			["r7", "james_liu", "alex_kim", "2023-06", "supervisor", "Postdoc supervision", "positive", 0.65],
			["r8", "sarah_martinez", "tom_rivera", "2023-01", "supervisor", "Lab oversight", "neutral", 0.55],
			["r9", "maya_patel", "alex_kim", "2024-01", "collaborator", "Joint analysis work", "positive", 0.7],
			["r10", "james_liu", "tom_rivera", "2023-06", "supervisor", "Technical guidance", "neutral", 0.45]
		]
		:put relationship {id, from_person, to_person, timestamp => relationship_type, description, sentiment, strength}`,

		`?[id, person_id, timestamp, behavior_type, description] <- [
			["b1", "sarah_martinez", "2023-03", "mentoring", "Guided Maya on research methods"],
			["b2", "sarah_martinez", "2023-11", "publishing", "Published paper in Nature"],
			["b3", "james_liu", "2023-11", "publishing", "Co-authored Nature paper"],
			["b4", "maya_patel", "2023-09", "presenting", "Conference presentation"],
			["b5", "sarah_martinez", "2024-01", "grant_writing", "Wrote AI safety grant proposal"],
			["b6", "alex_kim", "2024-01", "collaborating", "Increased cross-team collaboration"],
			["b7", "james_liu", "2023-06", "mentoring", "Onboarded new postdoc"],
			["b8", "tom_rivera", "2023-09", "technical_support", "Set up new lab equipment"]
		]
		:put behavior {id, person_id, timestamp => behavior_type, description}`,

		`?[id, timestamp, description] <- [
			["e1", "2023-01", "Research lab established with initial team"],
			["e2", "2023-06", "New team member Alex Kim joined"],
			["e3", "2023-09", "Conference presentation by team"],
			["e4", "2023-11", "Paper published in Nature"],
			["e5", "2024-02", "Grant awarded for AI safety research"]
		]
		:put event {id, timestamp => description}`,
	}
	for _, s := range seeds {
		if _, err := db.ExecScript(ctx, s, nil, nil); err != nil {
			return fmt.Errorf("seed data: %w", err)
		}
	}

	return nil
}
