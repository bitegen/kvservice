-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions (
    sequence BIGSERIAL PRIMARY KEY, 
    event_type SMALLINT, 
    key TEXT, 
    value TEXT 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE transactions;
-- +goose StatementEnd
