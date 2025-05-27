-- Drop composite indexes
DROP INDEX IF EXISTS idx_actions_user_note;
DROP INDEX IF EXISTS idx_notes_user_date;

-- Drop user_id indexes
DROP INDEX IF EXISTS idx_actions_user_id;
DROP INDEX IF EXISTS idx_notes_user_id;

-- Remove user_id columns
ALTER TABLE actions DROP COLUMN IF EXISTS user_id;
ALTER TABLE notes DROP COLUMN IF EXISTS user_id; 