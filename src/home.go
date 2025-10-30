package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
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

	db, err := GetDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.QueryContext(ctx, `
SELECT
    t.id,
    t.user_id,
    t.body,
    t.likes,
    t.saves,
    t.restacks,
    t.replies,
    t.is_edited,
    t.created_at,
    t.last_edited_at,
    u.id,
    u.created_at,
    u.username,
    u.profile_name,
    u.profile_url,
    u.bio
FROM tweets t
JOIN users u ON u.id = t.user_id
ORDER BY t.created_at DESC, t.id DESC
LIMIT 10
`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tweets []TweetWithUser
	for rows.Next() {
		tweet, err := scanTweetWithUser(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tweets = append(tweets, tweet)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tweets)
	log.Println("sent successfully")
}
