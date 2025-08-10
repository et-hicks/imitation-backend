package models

import "time"

// UserTweetInteraction represents per-user interactions with a tweet.
// Mirrors table public.user_tweet_interactions.
type UserTweetInteraction struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	TweetID     *int      `json:"tweet_id"`
	CommentID   *int      `json:"comment_id"`
	IsSaved     bool      `json:"is_saved"`
	IsLiked     bool      `json:"is_liked"`
	IsRestacked bool      `json:"is_restacked"`
	CreatedAt   time.Time `json:"created_at"`
}
