-- Migration: Fix settings table timestamps
-- Created: 2025-11-06
-- Purpose: Correct timestamp format in settings table

UPDATE settings
SET updated_at = datetime('now')
WHERE id = 'singleton';
