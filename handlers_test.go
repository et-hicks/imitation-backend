package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	api "github.com/et-hicks/imitation-backend/src"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := "file:testdb_" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	if err := os.Setenv("DATABASE_PATH", dsn); err != nil {
		t.Fatalf("set env: %v", err)
	}
	api.ResetDatabaseForTests()
	db, err := api.GetDB(context.Background())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	cleanupTables(t, db)
	return db
}

func cleanupTables(t *testing.T, db *sql.DB) {
	t.Helper()
	tables := []string{"user_following", "user_tweet_interactions", "comments", "tweets", "users"}
	for _, tbl := range tables {
		if _, err := db.Exec("DELETE FROM " + tbl); err != nil {
			t.Fatalf("cleanup %s: %v", tbl, err)
		}
	}
}

func insertUser(t *testing.T, db *sql.DB, id int) {
	t.Helper()
	hash, err := api.HashPasswordForTests("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (id, username, password_hash, profile_name, profile_url, bio) VALUES (?, ?, ?, ?, ?, ?)`,
		id, fmt.Sprintf("user%d", id), hash, fmt.Sprintf("User %d", id), "", "")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

func insertTweet(t *testing.T, db *sql.DB, id, userID int, body string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO tweets (id, user_id, body) VALUES (?, ?, ?)`, id, userID, body)
	if err != nil {
		t.Fatalf("insert tweet: %v", err)
	}
}

func seedUsersAndTweets(t *testing.T, db *sql.DB, userCount, tweetsPerUser int) {
	t.Helper()
	tweetID := 1
	for user := 1; user <= userCount; user++ {
		insertUser(t, db, user)
		for i := 0; i < tweetsPerUser; i++ {
			insertTweet(t, db, tweetID, user, fmt.Sprintf("User %d tweet %d", user, i+1))
			tweetID++
		}
	}
}

func TestHomeReturnsTen(t *testing.T) {
	db := setupTestDB(t)
	for i := 1; i <= 10; i++ {
		insertUser(t, db, i)
		insertTweet(t, db, i, i, fmt.Sprintf("Body %d", i))
	}

	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 10 {
		t.Fatalf("want 10 rows, got %d", len(got))
	}
}

func TestUser10ReturnsTenWithUserID10(t *testing.T) {
	db := setupTestDB(t)
	insertUser(t, db, 10)
	for i := 0; i < 10; i++ {
		insertTweet(t, db, 100+i, 10, "Body")
	}

	req := httptest.NewRequest(http.MethodGet, "/user/10", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 10 {
		t.Fatalf("want 10 rows, got %d", len(got))
	}
	for i, row := range got {
		if v, ok := row["user_id"].(float64); !ok || int(v) != 10 {
			t.Fatalf("row %d: want user_id=10, got %v", i, row["user_id"])
		}
	}
}

func TestTweet1HasExpectedFields(t *testing.T) {
	db := setupTestDB(t)
	insertUser(t, db, 1)
	body := "Tech company unveils new AI chip to speed up machine learning."
	insertTweet(t, db, 1, 1, body)

	req := httptest.NewRequest(http.MethodGet, "/tweet/1", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v, ok := got["id"].(float64); !ok || int(v) != 1 {
		t.Fatalf("want id=1, got %v", got["id"])
	}
	if v, ok := got["user_id"].(float64); !ok || int(v) != 1 {
		t.Fatalf("want user_id=1, got %v", got["user_id"])
	}
	if s, ok := got["body"].(string); !ok || s != body {
		t.Fatalf("unexpected body: %v", got["body"])
	}
}

func TestCreateCommentRequiresParentHeader(t *testing.T) {
	db := setupTestDB(t)
	insertUser(t, db, 1)

	body := bytes.NewBufferString(`{"body":"hi","is_comment":true}`)
	req := httptest.NewRequest(http.MethodPost, "/tweet", body)
	req.Header.Set("Authorization", "1")
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateCommentValidatesUser(t *testing.T) {
	setupTestDB(t)

	body := bytes.NewBufferString(`{"body":"hi","is_comment":true}`)
	req := httptest.NewRequest(http.MethodPost, "/tweet", body)
	req.Header.Set("Authorization", "999")
	req.Header.Set("Parent-Tweet-ID", "42")
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateCommentSuccess(t *testing.T) {
	db := setupTestDB(t)
	insertUser(t, db, 1)
	insertTweet(t, db, 42, 1, "Parent")

	body := bytes.NewBufferString(`{"body":"hi","is_comment":true}`)
	req := httptest.NewRequest(http.MethodPost, "/tweet", body)
	req.Header.Set("Authorization", "1")
	req.Header.Set("Parent-Tweet-ID", "42")
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v, ok := got["tweet_id"].(float64); !ok || int(v) != 42 {
		t.Fatalf("expected tweet_id=42, got %v", got["tweet_id"])
	}
}

func TestSignupHashesPassword(t *testing.T) {
	db := setupTestDB(t)

	body := bytes.NewBufferString(`{"username":"newuser","password":"super-secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", body)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	var hash string
	if err := db.QueryRow(`SELECT password_hash FROM users WHERE username = ?`, "newuser").Scan(&hash); err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	if hash == "" {
		t.Fatalf("password hash should not be empty")
	}
	if hash == "super-secret" {
		t.Fatalf("password hash stored in plaintext")
	}
	if !api.VerifyPasswordForTests(hash, "super-secret") {
		t.Fatalf("stored hash does not validate password")
	}
}

func TestLoginValidatesPassword(t *testing.T) {
	db := setupTestDB(t)

	hash, err := api.HashPasswordForTests("letmein")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users (username, password_hash) VALUES (?, ?)`, "login_user", hash); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	body := bytes.NewBufferString(`{"username":"login_user","password":"letmein"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
}

func TestLikeAuthCheck(t *testing.T) {
	db := setupTestDB(t)
	insertUser(t, db, 1)
	insertTweet(t, db, 10, 1, "Body")

	req := httptest.NewRequest(http.MethodPut, "/like/2/10", nil)
	req.Header.Set("Authorization", "1")
	req.Header.Set("Is-Comment", "false")
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPut, "/like/1/10", nil)
	req2.Header.Set("Authorization", "1")
	req2.Header.Set("Is-Comment", "false")
	rr2 := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestUnlikeAndUnsave(t *testing.T) {
	db := setupTestDB(t)
	insertUser(t, db, 1)
	insertTweet(t, db, 10, 1, "Body")

	req := httptest.NewRequest(http.MethodPut, "/like/1/10?remove=true", nil)
	req.Header.Set("Authorization", "1")
	req.Header.Set("Is-Comment", "false")
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPut, "/save/1/10?remove=true", nil)
	req2.Header.Set("Authorization", "1")
	rr2 := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestFollowAuthCheck(t *testing.T) {
	db := setupTestDB(t)
	insertUser(t, db, 1)
	insertUser(t, db, 3)

	req := httptest.NewRequest(http.MethodPut, "/follow/2/3", nil)
	req.Header.Set("Authorization", "1")
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPut, "/follow/1/3", nil)
	req2.Header.Set("Authorization", "1")
	rr2 := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestUserTweetsQueryAndEndpoint(t *testing.T) {
	db := setupTestDB(t)
	seedUsersAndTweets(t, db, 5, 5)

	userID := 3

	rows, err := db.Query(api.UserTweetsSelect, userID)
	if err != nil {
		t.Fatalf("query tweets: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tweets: %v", err)
	}
	if count != 5 {
		t.Fatalf("want 5 tweets from SQL, got %d", count)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/3/tweets", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	var got []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("want 5 tweets from endpoint, got %d", len(got))
	}
	for i, row := range got {
		if v, ok := row["user_id"].(float64); !ok || int(v) != userID {
			t.Fatalf("row %d: want user_id=%d, got %v", i, userID, row["user_id"])
		}
	}
}

func TestUserBookmarksQueryAndEndpoint(t *testing.T) {
	db := setupTestDB(t)
	seedUsersAndTweets(t, db, 5, 5)

	userID := 1
	firstForeignTweetID := 6 // tweets from user 2
	for i := 0; i < 5; i++ {
		tweetID := firstForeignTweetID + i
		if _, err := db.Exec(`INSERT INTO user_tweet_interactions (user_id, tweet_id, is_saved) VALUES (?, ?, 1)`, userID, tweetID); err != nil {
			t.Fatalf("insert bookmark: %v", err)
		}
	}

	rows, err := db.Query(api.UserBookmarkedTweetsSelect, userID)
	if err != nil {
		t.Fatalf("query bookmarks: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate bookmarks: %v", err)
	}
	if count != 5 {
		t.Fatalf("want 5 bookmarks from SQL, got %d", count)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/1/bookmarks", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	var got []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("want 5 bookmarks from endpoint, got %d", len(got))
	}
}

func TestUserLikesQueryAndEndpoint(t *testing.T) {
	db := setupTestDB(t)
	seedUsersAndTweets(t, db, 5, 5)

	userID := 2
	startTweetID := 11 // tweets from user 3
	for i := 0; i < 5; i++ {
		tweetID := startTweetID + i
		if _, err := db.Exec(`INSERT INTO user_tweet_interactions (user_id, tweet_id, is_liked) VALUES (?, ?, 1)`, userID, tweetID); err != nil {
			t.Fatalf("insert like: %v", err)
		}
	}

	rows, err := db.Query(api.UserLikedTweetsSelect, userID)
	if err != nil {
		t.Fatalf("query likes: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate likes: %v", err)
	}
	if count != 5 {
		t.Fatalf("want 5 likes from SQL, got %d", count)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/2/likes", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	var got []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("want 5 likes from endpoint, got %d", len(got))
	}
}
