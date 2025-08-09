package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	postgrest "github.com/supabase-community/postgrest-go"
)

func init() {
	http.HandleFunc("/home", homeHandler)
}

// homeHandler returns the 10 most recent tweets with user information.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("inilizied request")
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := GetSupabase(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tweets []TweetWithUser
	qb := client.From("tweets").Select("*,users(*)", "", false)
	qb = qb.Order("created_at", &postgrest.OrderOpts{Ascending: false})
	qb = qb.Limit(10, "")
	if _, err := qb.ExecuteTo(&tweets); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tweets)
	log.Println("sent successfully")
}
