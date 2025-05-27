-- Add user_id column to notes table
ALTER TABLE notes ADD COLUMN user_id VARCHAR(255) NOT NULL DEFAULT '';

-- Add user_id column to actions table  
ALTER TABLE actions ADD COLUMN user_id VARCHAR(255) NOT NULL DEFAULT '';

-- Create indexes for user_id columns for better query performance
CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_actions_user_id ON actions(user_id);

-- Create composite indexes for common query patterns
CREATE INDEX idx_notes_user_date ON notes(user_id, date);
CREATE INDEX idx_actions_user_note ON actions(user_id, note_id); 