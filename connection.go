package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	supabase "github.com/supabase-community/supabase-go"
)

// Package-level Supabase client initialized on first use.
var (
	sbClient *supabase.Client
	sbOnce   sync.Once
	sbErr    error
)

// getSupabase initializes and returns a shared Supabase client.
// Requires SUPABASE_URL and SUPABASE_KEY to be set.
func getSupabase(ctx context.Context) (*supabase.Client, error) {
	sbOnce.Do(func() {
		url := os.Getenv("SUPABASE_URL")
		key := os.Getenv("SUPABASE_KEY")
		if url == "" || key == "" {
			sbErr = fmt.Errorf("SUPABASE_URL or SUPABASE_KEY not set")
			return
		}
		sbClient, sbErr = supabase.NewClient(url, key, nil)
	})
	if sbErr != nil {
		return nil, sbErr
	}
	return sbClient, nil
}

// QueryTableAllRows fetches all rows from the given table via Supabase PostgREST
// and returns them as a slice of column-name->value maps.
func QueryTableAllRows(ctx context.Context, table string) ([]map[string]any, error) {
	client, err := getSupabase(ctx)
	if err != nil {
		return nil, err
	}
	data, _, err := client.From(table).Select("*", "", false).Execute()
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}
