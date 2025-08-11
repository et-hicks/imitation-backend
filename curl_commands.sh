#!/usr/bin/env bash

# Base URL of the server
BASE_URL="${BASE_URL:-http://localhost:8080}"

# Sample IDs used in requests; replace as needed
user_id="123"
tweet_id="1"
follow_id="456"

# --------------------
# Root endpoints
# --------------------

# Get index page
curl -X GET "$BASE_URL/"

# Get about page
curl -X GET "$BASE_URL/about"

# Get sample data
curl -X GET "$BASE_URL/data"

# --------------------
# API endpoints
# --------------------

# Home timeline
curl -X GET "$BASE_URL/home"

# Fetch a specific tweet
curl -X GET "$BASE_URL/tweet/$tweet_id"

# Fetch comments for a tweet
curl -X GET "$BASE_URL/tweet/$tweet_id/comments"

# Create a new tweet
# Replace {user_id} and the body content as needed
curl -X POST "$BASE_URL/tweet" \
  -H "Content-Type: application/json" \
  -H "Authorization: $user_id" \
  -d '{"body": "Hello world", "is_comment": false}'

# Create a comment on a tweet
curl -X POST "$BASE_URL/tweet" \
  -H "Content-Type: application/json" \
  -H "Authorization: $user_id" \
  -H "Parent-Tweet-ID: $tweet_id" \
  -d '{"body": "Nice post!", "is_comment": true}'

# Get the latest tweets for a user
curl -X GET "$BASE_URL/user/$user_id"

# Update a user's bio
curl -X POST "$BASE_URL/user/$user_id/bio" \
  -H "Content-Type: application/json" \
  -d '{"bio": "New bio"}'

# Like a tweet
curl -X PUT "$BASE_URL/like/$user_id/$tweet_id" \
  -H "Authorization: $user_id" \
  -H "Is-Comment: false"

# Remove like from a tweet
curl -X PUT "$BASE_URL/like/$user_id/$tweet_id?remove=true" \
  -H "Authorization: $user_id" \
  -H "Is-Comment: false"

# Save a tweet
curl -X PUT "$BASE_URL/save/$user_id/$tweet_id" \
  -H "Authorization: $user_id"

# Remove a saved tweet
curl -X PUT "$BASE_URL/save/$user_id/$tweet_id?remove=true" \
  -H "Authorization: $user_id"

# Restack a tweet
curl -X PUT "$BASE_URL/restack/$user_id/$tweet_id" \
  -H "Authorization: $user_id"

# Follow a user
curl -X PUT "$BASE_URL/follow/$user_id/$follow_id" \
  -H "Authorization: $user_id"
