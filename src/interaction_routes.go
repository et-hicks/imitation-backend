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
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client, err := GetSupabase(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payload := map[string]interface{}{
		"user_id":  userID,
		"is_liked": true,
	}
	if isComment {
		payload["comment_id"] = targetID
	} else {
		payload["tweet_id"] = targetID
	}
	qb := client.From("user_tweet_interactions").Insert(payload, true, "user_id,tweet_id,comment_id", "", "")
	if _, _, err := qb.Execute(); err != nil {
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
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client, err := GetSupabase(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payload := map[string]interface{}{
		"user_id":  userID,
		"tweet_id": tweetID,
		"is_saved": true,
	}
	qb := client.From("user_tweet_interactions").Insert(payload, true, "user_id,tweet_id,comment_id", "", "")
	if _, _, err := qb.Execute(); err != nil {
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
	client, err := GetSupabase(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payload := map[string]interface{}{
		"user_id":      userID,
		"tweet_id":     tweetID,
		"is_restacked": true,
	}
	qb := client.From("user_tweet_interactions").Insert(payload, true, "user_id,tweet_id,comment_id", "", "")
	if _, _, err := qb.Execute(); err != nil {
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
	client, err := GetSupabase(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payload := map[string]interface{}{
		"user_id":           userID,
		"following_user_id": followID,
	}
	qb := client.From("user_following").Insert(payload, true, "user_id,following_user_id", "", "")
	if _, _, err := qb.Execute(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Println("sent successfully")
}
