-- Migration 038: Fix Missing normalized_data Column
-- Migration 028 was supposed to add this but it's missing
-- This migration adds the missing normalized_data column to events table

ALTER TABLE events ADD COLUMN normalized_data TEXT;
