package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	tid, err := strconv.Atoi(tweetID)
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

	row := db.QueryRowContext(ctx, `
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
WHERE t.id = ?
`, tid)

	tweet, err := scanTweetWithUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
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
	tid, err := strconv.Atoi(tweetID)
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

	rows, err := db.QueryContext(ctx, `
SELECT
    c.id,
    c.user_id,
    c.tweet_id,
    c.body,
    c.likes,
    c.replies,
    c.is_edited,
    c.last_edited_at,
    c.created_at,
    u.id,
    u.created_at,
    u.username,
    u.profile_name,
    u.profile_url,
    u.bio
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.tweet_id = ?
ORDER BY c.created_at DESC, c.id DESC
`, tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []CommentWithUser
	for rows.Next() {
		comment, err := scanCommentWithUser(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
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

	if strings.TrimSpace(payload.Body) == "" {
		http.Error(w, "body is required", http.StatusBadRequest)
		return
	}

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

	db, err := GetDB(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := ensureUserExists(ctx, db, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if payload.IsComment {
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

		res, err := db.ExecContext(ctx, "INSERT INTO comments (user_id, tweet_id, body) VALUES (?, ?, ?)", userID, parentID, payload.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		commentID, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		row := db.QueryRowContext(ctx, `
SELECT
    id,
    user_id,
    tweet_id,
    body,
    likes,
    replies,
    is_edited,
    last_edited_at,
    created_at
FROM comments
WHERE id = ?
`, commentID)
		comment, err := scanComment(row)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(comment)
		log.Println("sent successfully")
		return
	}

	res, err := db.ExecContext(ctx, "INSERT INTO tweets (user_id, body) VALUES (?, ?)", userID, payload.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tweetIDVal, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	row := db.QueryRowContext(ctx, `
SELECT
    id,
    user_id,
    body,
    likes,
    saves,
    restacks,
    replies,
    is_edited,
    created_at,
    last_edited_at
FROM tweets
WHERE id = ?
`, tweetIDVal)
	tweet, err := scanTweet(row)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tweet)
	log.Println("sent successfully")
}

func ensureUserExists(ctx context.Context, db *sql.DB, userID int) error {
	var id int
	return db.QueryRowContext(ctx, "SELECT id FROM users WHERE id = ?", userID).Scan(&id)
}
