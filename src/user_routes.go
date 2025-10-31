package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	http.HandleFunc("/user/", userHandler)
	http.HandleFunc("/users/", usersHandler)
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

	uid, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	db, err := GetDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tweets, err := fetchTweets(ctx, db, UserTweetsSelect+"\nLIMIT 10", uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tweets)
	log.Println("sent successfully")
}

// usersHandler dispatches requests for plural user resources.
func usersHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}

	if parts[0] != "users" {
		http.NotFound(w, r)
		return
	}

	uid, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	switch parts[2] {
	case "tweets":
		handleUserTweets(w, r, uid)
	case "likes":
		handleUserLikes(w, r, uid)
	case "bookmarks":
		handleUserBookmarks(w, r, uid)
	default:
		http.NotFound(w, r)
	}
}

func handleUserTweets(w http.ResponseWriter, r *http.Request, userID int) {
	tweets, err := fetchTweetsForUser(r.Context(), UserTweetsSelect, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeTweetsResponse(w, tweets)
}

func handleUserLikes(w http.ResponseWriter, r *http.Request, userID int) {
	tweets, err := fetchTweetsForUser(r.Context(), UserLikedTweetsSelect, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeTweetsResponse(w, tweets)
}

func handleUserBookmarks(w http.ResponseWriter, r *http.Request, userID int) {
	tweets, err := fetchTweetsForUser(r.Context(), UserBookmarkedTweetsSelect, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeTweetsResponse(w, tweets)
}

func fetchTweetsForUser(ctx context.Context, query string, userID int) ([]TweetWithUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := GetDB(ctx)
	if err != nil {
		return nil, err
	}

	return fetchTweets(ctx, db, query, userID)
}

func fetchTweets(ctx context.Context, db *sql.DB, query string, args ...any) ([]TweetWithUser, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []TweetWithUser
	for rows.Next() {
		tweet, err := scanTweetWithUser(rows)
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, tweet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tweets, nil
}

func writeTweetsResponse(w http.ResponseWriter, tweets []TweetWithUser) {
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

	uid, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	db, err := GetDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := db.ExecContext(ctx, "UPDATE users SET bio = ? WHERE id = ?", payload.Bio, uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count, err := res.RowsAffected(); err == nil && count == 0 {
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Println("sent successfully")
}
