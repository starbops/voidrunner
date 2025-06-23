ALTER TABLE tasks 
ADD COLUMN user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE;

-- Create index for better query performance
CREATE INDEX idx_tasks_user_id ON tasks(user_id);