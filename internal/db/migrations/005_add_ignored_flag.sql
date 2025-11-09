-- Add ignored flag to unresolved_identities
-- This allows users to ignore unresolved identities (bots, external contributors, test accounts)
-- without permanently deleting them from the database

ALTER TABLE unresolved_identities ADD COLUMN ignored BOOLEAN NOT NULL DEFAULT 0;

-- Add index for filtering ignored identities
CREATE INDEX idx_unresolved_ignored ON unresolved_identities(ignored);
