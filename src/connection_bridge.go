package api

import (
	"context"
	"fmt"
	"os"
	"sync"

	supabase "github.com/supabase-community/supabase-go"
)

var (
	sbClientSrc *supabase.Client
	sbOnceSrc   sync.Once
	sbErrSrc    error
)

// GetSupabase provides a Supabase client for files under src/.
func GetSupabase(ctx context.Context) (*supabase.Client, error) {
	sbOnceSrc.Do(func() {
		url := os.Getenv("SUPABASE_URL")
		key := os.Getenv("SUPABASE_KEY")
		if url == "" || key == "" {
			sbErrSrc = fmt.Errorf("SUPABASE_URL or SUPABASE_KEY not set")
			return
		}
		sbClientSrc, sbErrSrc = supabase.NewClient(url, key, nil)
	})
	if sbErrSrc != nil {
		return nil, sbErrSrc
	}
	return sbClientSrc, nil
}

// ResetSupabaseForTests clears the cached client for tests.
func ResetSupabaseForTests() {
	sbClientSrc = nil
	sbErrSrc = nil
	sbOnceSrc = sync.Once{}
}
