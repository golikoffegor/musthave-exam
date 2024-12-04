-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 't_action') THEN
        CREATE TYPE "t_action" AS ENUM ('Withdraw', 'Debit');
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 't_status') THEN
        CREATE TYPE "t_status" AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS t_status;
DROP TYPE IF EXISTS t_action;
-- +goose StatementEnd
