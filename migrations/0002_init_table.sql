-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "user" (
	id bigint GENERATED ALWAYS AS IDENTITY NOT NULL,
	login varchar NOT NULL,
	password text NOT NULL,
	balance numeric NULL,
	CONSTRAINT user_pk PRIMARY KEY (id),
	CONSTRAINT user_unique UNIQUE (login)
);
CREATE TABLE IF NOT EXISTS transactions (
	id numeric NOT NULL,
	user_id bigint NULL,
	summ numeric NULL,
	date timestamp NULL,
	status t_status NOT NULL,
	action t_action NOT NULL,
	CONSTRAINT transactions_user_fk_1 FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
	CONSTRAINT transactions_id_unique UNIQUE (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE transactions;
DROP TABLE user;
-- +goose StatementEnd
