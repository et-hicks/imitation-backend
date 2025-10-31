package api

const (
	// UserTweetsSelect lists tweets authored by the provided user ordered from newest to oldest.
	UserTweetsSelect = `
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
WHERE t.user_id = ?
ORDER BY t.created_at DESC, t.id DESC`

	// UserLikedTweetsSelect lists tweets liked by the provided user ordered from newest like target.
	UserLikedTweetsSelect = `
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
FROM user_tweet_interactions uti
JOIN tweets t ON t.id = uti.tweet_id
JOIN users u ON u.id = t.user_id
WHERE uti.user_id = ? AND uti.is_liked = 1
ORDER BY t.created_at DESC, t.id DESC`

	// UserBookmarkedTweetsSelect lists tweets bookmarked by the provided user ordered from newest bookmark target.
	UserBookmarkedTweetsSelect = `
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
FROM user_tweet_interactions uti
JOIN tweets t ON t.id = uti.tweet_id
JOIN users u ON u.id = t.user_id
WHERE uti.user_id = ? AND uti.is_saved = 1
ORDER BY t.created_at DESC, t.id DESC`
)
