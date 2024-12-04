package repository

import (
	"context"
	"database/sql"
	"musthave-exam/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func (r *repo) RegisterUser(ctx context.Context, user model.User) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO "user" (login, password, balance) VALUES ($1, $2, $3) RETURNING id`
	var userID int64
	err = r.db.QueryRowContext(ctx, query, user.Login, hashedPassword, 0).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.WithError(err).Warning(model.ErrLoginAlreadyTaken.Error())
			return 0, model.ErrLoginAlreadyTaken
		}
		return 0, err
	}

	return userID, nil
}

func (r *repo) LoginUser(ctx context.Context, user model.User) (int64, error) {
	var storedUser model.User
	query := `SELECT id, password FROM "user" WHERE login=$1`
	err := r.db.QueryRowContext(ctx, query, user.Login).Scan(&storedUser.ID, &storedUser.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, model.ErrInvalidLoginPass
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		return 0, model.ErrInvalidLoginPass
	}

	return storedUser.ID, nil
}

func (r *repo) GetUser(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	query := `SELECT id, login, balance FROM "user" WHERE id=$1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Login, &user.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, model.ErrUserNotFound
		}
		return user, err
	}

	return user, nil
}

func (r *repo) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	var mbalance model.Balance
	var balance sql.NullFloat64
	query := `SELECT balance FROM "user" WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		r.log.WithError(err).Error("GetBalance:Failed to get balance")
		return mbalance, err
	}

	var withdrawn sql.NullFloat64
	queryWithdrawn := `SELECT SUM(summ) FROM transactions WHERE user_id = $1 AND action = 'Withdraw'`
	err = r.db.QueryRowContext(ctx, queryWithdrawn, userID).Scan(&withdrawn)
	if err != nil {
		r.log.WithError(err).Error("Failed to get withdrawn amount")
		return mbalance, err
	}

	if balance.Valid {
		mbalance.Current = balance.Float64
	}

	if withdrawn.Valid {
		mbalance.Withdrawn = withdrawn.Float64
	}

	return mbalance, nil
}
