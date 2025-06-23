DROP INDEX IF EXISTS idx_tasks_user_id;
ALTER TABLE tasks DROP COLUMN user_id;