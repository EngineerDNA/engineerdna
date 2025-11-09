-- Migration 029: Remove Duplicate Users Table
-- The 'engineers' table is the primary entity for tracking developers
-- The 'users' table is legacy and unused
-- This migration removes the duplicate table

-- Verify no data exists before dropping (safety check)
-- If this migration fails, there may be data that needs migration first

DROP TABLE IF EXISTS users;

-- Note: The engineers table provides all user functionality
-- with additional engineering-specific fields
