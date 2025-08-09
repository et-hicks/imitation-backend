package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	postgrest "github.com/supabase-community/postgrest-go"
)

func init() {
	http.HandleFunc("/user/", userHandler)
}

// userHandler dispatches user related routes.
func userHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	id := parts[1]

	if len(parts) == 2 && r.Method == http.MethodGet {
		userTweets(w, r, id)
		return
	}

	if len(parts) == 3 && parts[2] == "bio" && r.Method == http.MethodPost {
		updateBio(w, r, id)
		return
	}

	http.NotFound(w, r)
}

// userTweets returns 10 latest tweets for the specified user.
func userTweets(w http.ResponseWriter, r *http.Request, userID string) {
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
	qb = qb.Eq("user_id", userID)
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

// updateBio updates the bio for a given user.
func updateBio(w http.ResponseWriter, r *http.Request, userID string) {
	log.Println("inilizied request")
	var payload struct {
		Bio string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := GetSupabase(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qb := client.From("users").Update(map[string]string{"bio": payload.Bio}, "", "")
	qb = qb.Eq("id", userID)
	if _, _, err := qb.Execute(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Println("sent successfully")
}
