package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
		Body      string `json:"body"`
		IsComment bool   `json:"is_comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve user auth information from headers
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(authHeader)
	if err != nil {
		http.Error(w, "invalid authorization", http.StatusUnauthorized)
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

	// When posting a comment, validate the user and require parent tweet ID
	if payload.IsComment {
		// Ensure parent tweet id is provided in headers
		parentIDStr := r.Header.Get("Parent-Tweet-ID")
		if parentIDStr == "" {
			http.Error(w, "missing parent tweet id", http.StatusBadRequest)
			return
		}
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil {
			http.Error(w, "invalid parent tweet id", http.StatusBadRequest)
			return
		}

		// Validate that the user exists in the database
		if _, _, err := client.From("users").Select("id", "", false).Eq("id", strconv.Itoa(userID)).Single().Execute(); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		qb := client.From("comments").Insert(map[string]interface{}{
			"user_id":  userID,
			"tweet_id": parentID,
			"body":     payload.Body,
		}, false, "", "", "")
		data, _, err := qb.Single().Execute()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var comment models.Comment
		if err := json.Unmarshal(data, &comment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(comment)
		log.Println("sent successfully")
		return
	}

	qb := client.From("tweets").Insert(map[string]interface{}{
		"user_id": userID,
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
