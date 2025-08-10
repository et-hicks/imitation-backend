-- Auth.js-compatible schema for Supabase Adapter
-- Creates next_auth schema, tables, grants, and helper function

-- Ensure UUID functions exist (used by Auth.js schema)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Schema and privileges
CREATE SCHEMA IF NOT EXISTS next_auth;
GRANT USAGE ON SCHEMA next_auth TO service_role;
GRANT ALL ON SCHEMA next_auth TO postgres;

-- next_auth.users
CREATE TABLE IF NOT EXISTS next_auth.users (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name text,
  email text,
  "emailVerified" timestamptz,
  image text,
  CONSTRAINT email_unique UNIQUE (email)
);
GRANT ALL ON TABLE next_auth.users TO postgres;
GRANT ALL ON TABLE next_auth.users TO service_role;

-- Helper to read current user id from JWT claims (for RLS)
CREATE OR REPLACE FUNCTION next_auth.uid() RETURNS uuid
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    NULLIF(current_setting('request.jwt.claim.sub', true), ''),
    (NULLIF(current_setting('request.jwt.claims', true), '')::jsonb ->> 'sub')
  )::uuid
$$;

-- next_auth.sessions
CREATE TABLE IF NOT EXISTS next_auth.sessions (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  expires timestamptz NOT NULL,
  "sessionToken" text NOT NULL UNIQUE,
  "userId" uuid REFERENCES next_auth.users(id) ON DELETE CASCADE
);
GRANT ALL ON TABLE next_auth.sessions TO postgres;
GRANT ALL ON TABLE next_auth.sessions TO service_role;

-- next_auth.accounts
CREATE TABLE IF NOT EXISTS next_auth.accounts (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  type text NOT NULL,
  provider text NOT NULL,
  "providerAccountId" text NOT NULL,
  refresh_token text,
  access_token text,
  expires_at bigint,
  token_type text,
  scope text,
  id_token text,
  session_state text,
  oauth_token_secret text,
  oauth_token text,
  "userId" uuid REFERENCES next_auth.users(id) ON DELETE CASCADE,
  CONSTRAINT provider_unique UNIQUE (provider, "providerAccountId")
);
GRANT ALL ON TABLE next_auth.accounts TO postgres;
GRANT ALL ON TABLE next_auth.accounts TO service_role;

-- next_auth.verification_tokens
CREATE TABLE IF NOT EXISTS next_auth.verification_tokens (
  identifier text,
  token text,
  expires timestamptz NOT NULL,
  PRIMARY KEY (token),
  CONSTRAINT token_identifier_unique UNIQUE (token, identifier)
);
GRANT ALL ON TABLE next_auth.verification_tokens TO postgres;
GRANT ALL ON TABLE next_auth.verification_tokens TO service_role;


