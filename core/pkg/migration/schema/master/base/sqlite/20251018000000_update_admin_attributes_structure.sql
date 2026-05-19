-- +goose Up
-- Update admin user's attributes to new nested structure
-- Only update if it's exactly the value set by 20250928000000_alter_table_user_role.sql
UPDATE users 
SET extra = '{"roles":{"role:super_admin":true},"info":{}}'
WHERE username = 'admin' 
  AND extra = '{"role:super_admin":true}';

-- +goose Down
-- Rollback to legacy flat structure
UPDATE users 
SET extra = '{"role:super_admin":true}'
WHERE username = 'admin'
  AND extra = '{"roles":{"role:super_admin":true},"info":{}}';
