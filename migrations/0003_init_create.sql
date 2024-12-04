-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_date()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_date_trigger
BEFORE INSERT ON transactions
FOR EACH ROW
EXECUTE FUNCTION set_date();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd