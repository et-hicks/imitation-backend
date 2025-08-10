CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    username TEXT NOT NULL UNIQUE,
    profile_name TEXT,
    profile_url TEXT,
    bio TEXT
);

CREATE TABLE IF NOT EXISTS tweets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    body TEXT NOT NULL,
    likes INTEGER NOT NULL DEFAULT 0,
    saves INTEGER NOT NULL DEFAULT 0,
    restacks INTEGER NOT NULL DEFAULT 0,
    replies INTEGER NOT NULL DEFAULT 0,
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_edited_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    tweet_id INTEGER NOT NULL REFERENCES tweets(id),
    body TEXT NOT NULL,
    likes INTEGER NOT NULL DEFAULT 0,
    replies INTEGER NOT NULL DEFAULT 0,
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    last_edited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);



-- Ensure tweets has a comments column
ALTER TABLE IF EXISTS tweets
ADD COLUMN IF NOT EXISTS comments INTEGER NOT NULL DEFAULT 0;

-- User Tweet Interactions: tracks saves, likes, restacks per user/tweet
CREATE TABLE IF NOT EXISTS user_tweet_interactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    tweet_id INTEGER REFERENCES tweets(id),
    comment_id INTEGER REFERENCES comments(id),
    is_saved BOOLEAN NOT NULL DEFAULT FALSE,
    is_liked BOOLEAN NOT NULL DEFAULT FALSE,
    is_restacked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure constraint: at least one of tweet_id or comment_id is non-null
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_schema = 'public'
      AND table_name = 'user_tweet_interactions'
      AND constraint_name = 'user_tweet_interactions_has_target'
  ) THEN
    ALTER TABLE public.user_tweet_interactions
      ADD CONSTRAINT user_tweet_interactions_has_target
      CHECK (tweet_id IS NOT NULL OR comment_id IS NOT NULL);
  END IF;
END$$;

-- If table existed previously with NOT NULL on tweet_id, relax it
ALTER TABLE IF EXISTS public.user_tweet_interactions
  ALTER COLUMN tweet_id DROP NOT NULL;
