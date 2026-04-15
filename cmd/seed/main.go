// Seed tool: applies SQL seed files from a directory in filename order.
// Usage:
//   go run ./cmd/seed/ -database <postgres_url> -path ./migrations/seeds up
//   go run ./cmd/seed/ -database <postgres_url> -path ./migrations/seeds down
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := flag.String("database", "", "PostgreSQL connection URL (required)")
	dir := flag.String("path", "./migrations/seeds", "Directory with seed SQL files")
	flag.Parse()

	direction := "up"
	if flag.NArg() > 0 {
		direction = flag.Arg(0)
	}

	if *dbURL == "" {
		log.Fatal("usage: seed -database <url> -path <dir> [up|down]")
	}
	if direction != "up" && direction != "down" {
		log.Fatalf("direction must be 'up' or 'down', got: %s", direction)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *dbURL)
	if err != nil {
		log.Fatalf("connect to DB: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping DB: %v", err)
	}

	suffix := fmt.Sprintf(".%s.sql", direction)
	entries, err := os.ReadDir(*dir)
	if err != nil {
		log.Fatalf("read dir %s: %v", *dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			files = append(files, filepath.Join(*dir, e.Name()))
		}
	}

	if direction == "down" {
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	} else {
		sort.Strings(files)
	}

	if len(files) == 0 {
		log.Printf("No seed files found in %s for direction=%s", *dir, direction)
		return
	}

	for _, f := range files {
		log.Printf("Applying %s ...", filepath.Base(f))
		content, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("read %s: %v", f, err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			log.Fatalf("exec %s: %v", f, err)
		}
		log.Printf("OK: %s", filepath.Base(f))
	}

	log.Printf("Seed %s completed: %d file(s) applied", direction, len(files))
}
