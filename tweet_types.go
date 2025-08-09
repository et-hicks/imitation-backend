package main

import (
	"github.com/fly-apps/go-example/models"
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
