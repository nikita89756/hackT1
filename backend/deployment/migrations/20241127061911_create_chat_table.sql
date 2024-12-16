-- +goose Up
-- +goose StatementBegin
CREATE TABLE chats (
    chat_id SERIAL PRIMARY KEY,
    user_id INT,
    chat_name VARCHAR(255),
    model_name VARCHAR(255),
    file_url VARCHAR(255),
    instruction VARCHAR(255),
    embending VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chats;
-- +goose StatementEnd
