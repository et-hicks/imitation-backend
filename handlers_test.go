package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	api "github.com/et-hicks/imitation-backend/src"
)

// fakeSupabaseServer returns a test server that mimics minimal Supabase REST endpoints used by handlers.
func fakeSupabaseServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Tweets collection
	mux.HandleFunc("/rest/v1/tweets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Single tweet by id
		if id := r.URL.Query().Get("id"); id != "" {
			// Expect format eq.1
			if id == "eq.1" {
				_, _ = w.Write([]byte(`{"id":1,"user_id":1,"body":"Tech company unveils new AI chip to speed up machine learning.","users":{"id":1}}`))
				return
			}
		}
		// Tweets filtered by user_id
		if uid := r.URL.Query().Get("user_id"); uid != "" {
			// Expect format eq.<n>
			if len(uid) > 3 && uid[:3] == "eq." {
				nStr := uid[3:]
				n, _ := strconv.Atoi(nStr)
				// Return 10 tweets for this user
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(genTweetsJSON(10, func(i int) (id, userID int, body string) {
					return i + 1, n, "Body"
				})))
				return
			}
		}
		// Default: return 10 tweets for /home
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(genTweetsJSON(10, func(i int) (id, userID int, body string) {
			return i + 1, i + 1, "Body"
		})))
	})

	// Users collection for auth validation
	mux.HandleFunc("/rest/v1/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if id := r.URL.Query().Get("id"); id == "eq.1" {
			_, _ = w.Write([]byte(`[{"id":1}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// Comments collection for posting comments
	mux.HandleFunc("/rest/v1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			// Ensure body is read so client isn't left hanging
			_, _ = io.ReadAll(r.Body)
			_, _ = w.Write([]byte(`{"id":1,"user_id":1,"tweet_id":42,"body":"test"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	// user_tweet_interactions for likes/saves/restacks
	mux.HandleFunc("/rest/v1/user_tweet_interactions", func(w http.ResponseWriter, r *http.Request) {
		// Just acknowledge the write
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	})

	// user_following for follow relationships
	mux.HandleFunc("/rest/v1/user_following", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	})

	return httptest.NewServer(mux)
}

// genTweetsJSON generates a JSON array of tweets with nested users.
func genTweetsJSON(n int, makeRow func(i int) (id, userID int, body string)) string {
	type U struct {
		ID int `json:"id"`
	}
	type T struct {
		ID     int    `json:"id"`
		UserID int    `json:"user_id"`
		Body   string `json:"body"`
		Users  U      `json:"users"`
	}
	rows := make([]T, 0, n)
	for i := 0; i < n; i++ {
		id, uid, body := makeRow(i)
		rows = append(rows, T{ID: id, UserID: uid, Body: body, Users: U{ID: uid}})
	}
	b, _ := json.Marshal(rows)
	return string(b)
}

// setSupabaseEnv points the handlers to the fake Supabase server.
func setSupabaseEnv(url string) {
	_ = os.Setenv("SUPABASE_URL", url)
	_ = os.Setenv("SUPABASE_KEY", "test-key")
}

func TestHomeReturnsTen(t *testing.T) {
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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
	if s, ok := got["body"].(string); !ok || s != "Tech company unveils new AI chip to speed up machine learning." {
		t.Fatalf("unexpected body: %v", got["body"])
	}
}

func TestCreateCommentRequiresParentHeader(t *testing.T) {
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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

func TestLikeAuthCheck(t *testing.T) {
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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
	srv := fakeSupabaseServer(t)
	defer srv.Close()
	setSupabaseEnv(srv.URL)
	api.ResetSupabaseForTests()

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
