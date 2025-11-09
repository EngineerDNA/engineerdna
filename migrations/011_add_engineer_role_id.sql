-- Add role_id to engineers table for scoring system
ALTER TABLE engineers ADD COLUMN role_id TEXT;
CREATE INDEX IF NOT EXISTS idx_engineers_role_id ON engineers(role_id);
