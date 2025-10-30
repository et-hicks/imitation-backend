package api

import (
	"database/sql"
	"strings"
	"time"

	"github.com/et-hicks/imitation-backend/models"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTweetWithUser(scanner rowScanner) (TweetWithUser, error) {
	var (
		tweet        models.Tweet
		user         models.User
		tweetCreated sql.NullString
		tweetEdited  sql.NullString
		userCreated  sql.NullString
		profile      sql.NullString
		profileURL   sql.NullString
		bio          sql.NullString
	)
	err := scanner.Scan(
		&tweet.ID,
		&tweet.UserID,
		&tweet.Body,
		&tweet.Likes,
		&tweet.Saves,
		&tweet.Restacks,
		&tweet.Replies,
		&tweet.IsEdited,
		&tweetCreated,
		&tweetEdited,
		&user.ID,
		&userCreated,
		&user.Username,
		&profile,
		&profileURL,
		&bio,
	)
	if err != nil {
		return TweetWithUser{}, err
	}
	tweet.CreatedAt = parseSQLiteTime(tweetCreated)
	tweet.LastEditedAt = parseSQLiteTime(tweetEdited)
	user.CreatedAt = parseSQLiteTime(userCreated)
	if profile.Valid {
		user.ProfileName = profile.String
	}
	if profileURL.Valid {
		user.ProfileURL = profileURL.String
	}
	if bio.Valid {
		user.Bio = bio.String
	}
	return TweetWithUser{Tweet: tweet, User: user}, nil
}

func scanCommentWithUser(scanner rowScanner) (CommentWithUser, error) {
	var (
		comment        models.Comment
		commentCreated sql.NullString
		commentEdited  sql.NullString
		user           models.User
		userCreated    sql.NullString
		profile        sql.NullString
		profileURL     sql.NullString
		bio            sql.NullString
	)
	err := scanner.Scan(
		&comment.ID,
		&comment.UserID,
		&comment.TweetID,
		&comment.Body,
		&comment.Likes,
		&comment.Replies,
		&comment.IsEdited,
		&commentEdited,
		&commentCreated,
		&user.ID,
		&userCreated,
		&user.Username,
		&profile,
		&profileURL,
		&bio,
	)
	if err != nil {
		return CommentWithUser{}, err
	}
	comment.CreatedAt = parseSQLiteTime(commentCreated)
	comment.LastEditedAt = parseSQLiteTime(commentEdited)
	user.CreatedAt = parseSQLiteTime(userCreated)
	if profile.Valid {
		user.ProfileName = profile.String
	}
	if profileURL.Valid {
		user.ProfileURL = profileURL.String
	}
	if bio.Valid {
		user.Bio = bio.String
	}
	return CommentWithUser{Comment: comment, User: user}, nil
}

func scanTweet(scanner rowScanner) (models.Tweet, error) {
	var (
		tweet        models.Tweet
		createdAt    sql.NullString
		lastEditedAt sql.NullString
	)
	err := scanner.Scan(
		&tweet.ID,
		&tweet.UserID,
		&tweet.Body,
		&tweet.Likes,
		&tweet.Saves,
		&tweet.Restacks,
		&tweet.Replies,
		&tweet.IsEdited,
		&createdAt,
		&lastEditedAt,
	)
	if err != nil {
		return models.Tweet{}, err
	}
	tweet.CreatedAt = parseSQLiteTime(createdAt)
	tweet.LastEditedAt = parseSQLiteTime(lastEditedAt)
	return tweet, nil
}

func scanComment(scanner rowScanner) (models.Comment, error) {
	var (
		comment      models.Comment
		createdAt    sql.NullString
		lastEditedAt sql.NullString
	)
	err := scanner.Scan(
		&comment.ID,
		&comment.UserID,
		&comment.TweetID,
		&comment.Body,
		&comment.Likes,
		&comment.Replies,
		&comment.IsEdited,
		&lastEditedAt,
		&createdAt,
	)
	if err != nil {
		return models.Comment{}, err
	}
	comment.CreatedAt = parseSQLiteTime(createdAt)
	comment.LastEditedAt = parseSQLiteTime(lastEditedAt)
	return comment, nil
}

func parseSQLiteTime(value sql.NullString) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	s := strings.TrimSpace(value.String)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, s); err == nil {
			return ts.UTC()
		}
	}
	return time.Time{}
}
