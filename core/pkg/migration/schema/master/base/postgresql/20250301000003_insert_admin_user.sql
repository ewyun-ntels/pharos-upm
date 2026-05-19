-- +goose Up
INSERT INTO users(username, password_hash) VALUES ('admin', '$2a$10$IX35uK8kCcEfKPmVkUYEoOmOJTDCUr/h3CQZHKmqKYTjilQTkzVwO')
ON CONFLICT(username)
DO UPDATE SET password_hash =' $2a$10$IX35uK8kCcEfKPmVkUYEoOmOJTDCUr/h3CQZHKmqKYTjilQTkzVwO';

-- +goose Down
DELETE FROM users WHERE username = 'admin';
