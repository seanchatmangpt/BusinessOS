-- Migration: 026_auth_invites.sql
-- Adds the auth_invites table for the team invite system (Tier 2 auth).
--
-- Design decisions:
-- • Only the SHA-256 hash of the raw token is stored. The raw token is sent
--   to the invitee once and never persisted, so a database breach cannot
--   reveal valid invite links.
-- • The `role` column mirrors the workspace RBAC roles and is applied to the
--   new user on registration.
-- • Indexes cover the common hot paths: token_hash lookup and admin dashboard
--   listing by invited_by / email.

CREATE TABLE IF NOT EXISTS auth_invites (
    id          VARCHAR(255) PRIMARY KEY,

    -- Invitee details
    email       VARCHAR(255) NOT NULL,
    role        VARCHAR(50)  NOT NULL DEFAULT 'member',  -- admin | member | viewer

    -- Security: only the SHA-256 hash is stored
    token_hash  VARCHAR(64)  NOT NULL UNIQUE,

    -- Tracking
    invited_by  VARCHAR(255) REFERENCES "user"(id) ON DELETE SET NULL,
    expires_at  TIMESTAMPTZ  NOT NULL,
    used_at     TIMESTAMPTZ,           -- NULL until the invite is consumed
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_invites_token_hash  ON auth_invites(token_hash);
CREATE INDEX IF NOT EXISTS idx_auth_invites_email       ON auth_invites(email);
CREATE INDEX IF NOT EXISTS idx_auth_invites_invited_by  ON auth_invites(invited_by);
CREATE INDEX IF NOT EXISTS idx_auth_invites_expires_at  ON auth_invites(expires_at);
