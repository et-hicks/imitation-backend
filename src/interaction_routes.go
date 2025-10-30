package api

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	http.HandleFunc("/like/", likeHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/restack/", restackHandler)
	http.HandleFunc("/follow/", followHandler)
}

func likeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}
	userIDStr, targetIDStr := parts[1], parts[2]
	if r.Header.Get("Authorization") != userIDStr {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	isCommentStr := r.Header.Get("Is-Comment")
	if isCommentStr == "" {
		http.Error(w, "missing Is-Comment header", http.StatusBadRequest)
		return
	}
	isComment := strings.ToLower(isCommentStr) == "true"
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "invalid target id", http.StatusBadRequest)
		return
	}
	remove := strings.ToLower(r.URL.Query().Get("remove")) == "true"
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, err := GetDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var query string
	var args []any
	if isComment {
		if remove {
			query = "INSERT INTO user_tweet_interactions (user_id, comment_id, is_liked) VALUES (?, ?, 0) ON CONFLICT(user_id, comment_id) DO UPDATE SET is_liked=0"
		} else {
			query = "INSERT INTO user_tweet_interactions (user_id, comment_id, is_liked) VALUES (?, ?, 1) ON CONFLICT(user_id, comment_id) DO UPDATE SET is_liked=1"
		}
		args = []any{userID, targetID}
	} else {
		if remove {
			query = "INSERT INTO user_tweet_interactions (user_id, tweet_id, is_liked) VALUES (?, ?, 0) ON CONFLICT(user_id, tweet_id) DO UPDATE SET is_liked=0"
		} else {
			query = "INSERT INTO user_tweet_interactions (user_id, tweet_id, is_liked) VALUES (?, ?, 1) ON CONFLICT(user_id, tweet_id) DO UPDATE SET is_liked=1"
		}
		args = []any{userID, targetID}
	}
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Println("sent successfully")
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}
	userIDStr, tweetIDStr := parts[1], parts[2]
	if r.Header.Get("Authorization") != userIDStr {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	tweetID, err := strconv.Atoi(tweetIDStr)
	if err != nil {
		http.Error(w, "invalid tweet id", http.StatusBadRequest)
		return
	}
	remove := strings.ToLower(r.URL.Query().Get("remove")) == "true"
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, err := GetDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var query string
	if remove {
		query = "INSERT INTO user_tweet_interactions (user_id, tweet_id, is_saved) VALUES (?, ?, 0) ON CONFLICT(user_id, tweet_id) DO UPDATE SET is_saved=0"
	} else {
		query = "INSERT INTO user_tweet_interactions (user_id, tweet_id, is_saved) VALUES (?, ?, 1) ON CONFLICT(user_id, tweet_id) DO UPDATE SET is_saved=1"
	}
	if _, err := db.ExecContext(ctx, query, userID, tweetID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Println("sent successfully")
}

func restackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}
	userIDStr, tweetIDStr := parts[1], parts[2]
	if r.Header.Get("Authorization") != userIDStr {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	tweetID, err := strconv.Atoi(tweetIDStr)
	if err != nil {
		http.Error(w, "invalid tweet id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, err := GetDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	query := "INSERT INTO user_tweet_interactions (user_id, tweet_id, is_restacked) VALUES (?, ?, 1) ON CONFLICT(user_id, tweet_id) DO UPDATE SET is_restacked=1"
	if _, err := db.ExecContext(ctx, query, userID, tweetID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Println("sent successfully")
}

func followHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}
	userIDStr, followIDStr := parts[1], parts[2]
	if r.Header.Get("Authorization") != userIDStr {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	followID, err := strconv.Atoi(followIDStr)
	if err != nil {
		http.Error(w, "invalid follow id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, err := GetDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	query := "INSERT INTO user_following (user_id, following_user_id) VALUES (?, ?) ON CONFLICT(user_id, following_user_id) DO NOTHING"
	if _, err := db.ExecContext(ctx, query, userID, followID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Println("sent successfully")
}
