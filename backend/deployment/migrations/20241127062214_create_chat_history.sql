-- +goose Up
-- +goose StatementBegin
CREATE TABLE history (
    message_id SERIAL PRIMARY KEY,
    chat_id INT NOT NULL ,
    user_id INT NOT NULL ,
    question TEXT NOT NULL,
    ai_answer TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE history;
-- +goose StatementEnd
