-- Migration 052: Add onboarding_completed tracking to users table
-- This allows us to redirect new users to onboarding flow

-- Add onboarding_completed column (defaults to false for new users)
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN DEFAULT FALSE;

-- Add index for quick lookups during login
CREATE INDEX IF NOT EXISTS idx_user_onboarding_completed ON "user"(onboarding_completed);

-- Mark existing users as onboarding completed (they already bypassed it)
UPDATE "user" SET onboarding_completed = TRUE WHERE onboarding_completed IS NULL OR onboarding_completed = FALSE;

-- Add comment
COMMENT ON COLUMN "user".onboarding_completed IS 'Whether user has completed OSA Build onboarding flow (13 screens)';
