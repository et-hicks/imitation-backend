package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/et-hicks/imitation-backend/models"
	postgrest "github.com/supabase-community/postgrest-go"
)

func init() {
	http.HandleFunc("/tweet", createTweet)
	http.HandleFunc("/tweet/", tweetHandler)
}

// tweetHandler handles retrieval of tweets and their comments.
func tweetHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	id := parts[1]

	if len(parts) == 2 && r.Method == http.MethodGet {
		fetchTweet(w, r, id)
		return
	}

	if len(parts) == 3 && parts[2] == "comments" && r.Method == http.MethodGet {
		fetchComments(w, r, id)
		return
	}

	http.NotFound(w, r)
}

// fetchTweet returns a specific tweet with user info.
func fetchTweet(w http.ResponseWriter, r *http.Request, tweetID string) {
	log.Println("inilizied request")
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := GetSupabase(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qb := client.From("tweets").Select("*,users(*)", "", false)
	qb = qb.Eq("id", tweetID)
	data, _, err := qb.Single().Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var tweet TweetWithUser
	if err := json.Unmarshal(data, &tweet); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tweet)
	log.Println("sent successfully")
}

// fetchComments returns comments for a tweet.
func fetchComments(w http.ResponseWriter, r *http.Request, tweetID string) {
	log.Println("inilizied request")
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := GetSupabase(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var comments []CommentWithUser
	qb := client.From("comments").Select("*,users(*)", "", false)
	qb = qb.Eq("tweet_id", tweetID)
	qb = qb.Order("created_at", &postgrest.OrderOpts{Ascending: false})
	if _, err := qb.ExecuteTo(&comments); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(comments)
	log.Println("sent successfully")
}

// createTweet inserts a new tweet for a user.
func createTweet(w http.ResponseWriter, r *http.Request) {
	log.Println("inilizied request")
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var payload struct {
		UserID int    `json:"user_id"`
		Body   string `json:"body"`
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

	qb := client.From("tweets").Insert(map[string]interface{}{
		"user_id": payload.UserID,
		"body":    payload.Body,
	}, false, "", "", "")
	data, _, err := qb.Single().Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tweet models.Tweet
	if err := json.Unmarshal(data, &tweet); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tweet)
	log.Println("sent successfully")
}
