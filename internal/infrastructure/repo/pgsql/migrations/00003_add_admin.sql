-- +goose Up
INSERT INTO users (login, password_hash)
VALUES('admin','$2a$10$blIZ78H6Xnt8ifCG6qsAY.MxelWlSsXRNFIpJfVC72aNLjWI9JNRK');

-- +goose Down
DELETE FROM users
WHERE login='admin';
