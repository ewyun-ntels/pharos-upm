-- +goose Up
INSERT OR REPLACE INTO users
       (username, password_hash)
       VALUES
       ('admin', '$2a$10$IX35uK8kCcEfKPmVkUYEoOmOJTDCUr/h3CQZHKmqKYTjilQTkzVwO');

-- +goose Down
DELETE FROM users WHERE username = 'admin';
