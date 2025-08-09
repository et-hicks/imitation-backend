package api

import (
	"github.com/et-hicks/imitation-backend/models"
)

// TweetWithUser combines tweet data with its author.
type TweetWithUser struct {
	models.Tweet
	User models.User `json:"users"`
}

// CommentWithUser combines comment data with its author.
type CommentWithUser struct {
	models.Comment
	User models.User `json:"users"`
}
