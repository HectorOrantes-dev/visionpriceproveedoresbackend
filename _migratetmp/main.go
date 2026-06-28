package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// readDatabaseURL parses DATABASE_URL out of the backend .env file.
func readDatabaseURL(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "DATABASE_URL") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			return v, nil
		}
	}
	return "", fmt.Errorf("DATABASE_URL not found")
}

func main() {
	apply := len(os.Args) > 1 && os.Args[1] == "--apply"

	url, err := readDatabaseURL(".env")
	if err != nil {
		fmt.Println("ERROR reading .env:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		fmt.Println("ERROR connecting:", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		fmt.Println("ERROR ping:", err)
		os.Exit(1)
	}
	fmt.Println("✅ Connected")

	printCols := func(stage string) {
		rows, err := pool.Query(ctx,
			`SELECT column_name, data_type FROM information_schema.columns
			 WHERE table_name = 'products' ORDER BY ordinal_position`)
		if err != nil {
			fmt.Println("ERROR query columns:", err)
			return
		}
		defer rows.Close()
		fmt.Printf("\n--- products columns (%s) ---\n", stage)
		for rows.Next() {
			var name, typ string
			_ = rows.Scan(&name, &typ)
			fmt.Printf("  %-14s %s\n", name, typ)
		}
	}

	printCols("before")

	if !apply {
		fmt.Println("\n(read-only; pass --apply to run the migration)")
		return
	}

	migration := `
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS sku    TEXT,
    ADD COLUMN IF NOT EXISTS brand  TEXT,
    ADD COLUMN IF NOT EXISTS stock  INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS status TEXT    NOT NULL DEFAULT 'active';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'products_status_check'
    ) THEN
        ALTER TABLE products
            ADD CONSTRAINT products_status_check
            CHECK (status IN ('active', 'draft', 'inactive', 'out_of_stock'));
    END IF;
END$$;
`
	if _, err := pool.Exec(ctx, migration); err != nil {
		fmt.Println("ERROR applying migration:", err)
		os.Exit(1)
	}
	fmt.Println("\n✅ Migration applied")
	printCols("after")
}
