-- +goose Up
DROP TABLE user_role;
DROP TABLE roles;

-- +goose Down
CREATE TABLE IF NOT EXISTS roles
(
    name      TEXT PRIMARY KEY NOT NULL,
    description TEXT NOT NULL DEFAULT ''
);

INSERT INTO roles (name, description) VALUES
    ('admin', 'Administrator with full access');

CREATE TABLE IF NOT EXISTS user_role
(
    username TEXT NOT NULL,
    role_name TEXT NOT NULL,
    PRIMARY KEY (username, role_name),
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE,
    FOREIGN KEY (role_name) REFERENCES roles(name) ON DELETE CASCADE
    );

INSERT INTO user_role (username, role_name) VALUES
    ('admin', 'admin');