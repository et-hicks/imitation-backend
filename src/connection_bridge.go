package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/et-hicks/imitation-backend/internal/sqlitedriver"
)

var (
	dbOnce sync.Once
	dbErr  error
	dbConn *sql.DB
)

const (
	defaultDBPath = "data/imitation.db"
	schemaSQL     = `PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username TEXT NOT NULL UNIQUE,
    profile_name TEXT,
    profile_url TEXT,
    bio TEXT
);

CREATE TABLE IF NOT EXISTS tweets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    likes INTEGER NOT NULL DEFAULT 0,
    saves INTEGER NOT NULL DEFAULT 0,
    restacks INTEGER NOT NULL DEFAULT 0,
    replies INTEGER NOT NULL DEFAULT 0,
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_edited_at DATETIME,
    comments INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tweet_id INTEGER NOT NULL REFERENCES tweets(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    likes INTEGER NOT NULL DEFAULT 0,
    replies INTEGER NOT NULL DEFAULT 0,
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    last_edited_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_tweet_interactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tweet_id INTEGER REFERENCES tweets(id) ON DELETE CASCADE,
    comment_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    is_saved BOOLEAN NOT NULL DEFAULT FALSE,
    is_liked BOOLEAN NOT NULL DEFAULT FALSE,
    is_restacked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (tweet_id IS NOT NULL OR comment_id IS NOT NULL),
    UNIQUE (user_id, tweet_id),
    UNIQUE (user_id, comment_id)
);

CREATE TABLE IF NOT EXISTS user_following (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, following_user_id)
);
`
)

// GetDB returns a shared SQLite connection for handlers under src/.
func GetDB(ctx context.Context) (*sql.DB, error) {
	dbOnce.Do(func() {
		var conn *sql.DB
		conn, dbErr = openDatabase()
		if dbErr != nil {
			return
		}
		// Verify the connection eagerly using the provided context.
		if err := pingContext(ctx, conn); err != nil {
			dbErr = err
			_ = conn.Close()
			return
		}
		if err := applySchema(ctx, conn); err != nil {
			dbErr = err
			_ = conn.Close()
			return
		}
		dbConn = conn
	})
	if dbErr != nil {
		return nil, dbErr
	}
	return dbConn, nil
}

func openDatabase() (*sql.DB, error) {
	rawPath := os.Getenv("DATABASE_PATH")
	if rawPath == "" {
		rawPath = defaultDBPath
	}
	dsn := rawPath
	if strings.HasPrefix(dsn, "file:") || strings.HasPrefix(dsn, ":memory:") {
		// pass through as provided to allow advanced URIs
	} else {
		if err := ensureDir(rawPath); err != nil {
			return nil, err
		}
		dsn = "file:" + rawPath
	}
	conn, err := sql.Open("custom_sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetConnMaxLifetime(0)
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	return conn, nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}
	return nil
}

func pingContext(ctx context.Context, db *sql.DB) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func applySchema(ctx context.Context, db *sql.DB) error {
	statements := strings.Split(schemaSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("run schema: %w", err)
		}
	}
	return nil
}

// ResetDatabaseForTests clears cached connections for tests.
func ResetDatabaseForTests() {
	if dbConn != nil {
		_ = dbConn.Close()
	}
	dbConn = nil
	dbErr = nil
	dbOnce = sync.Once{}
}
