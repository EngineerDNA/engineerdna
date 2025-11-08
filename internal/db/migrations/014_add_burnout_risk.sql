-- Add burnout risk tracking to performance scores
ALTER TABLE performance_scores ADD COLUMN burnout_risk TEXT;
