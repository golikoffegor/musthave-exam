package repository

import (
	"context"
	"database/sql"

	"github.com/golikoffegor/musthave-exam/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func (r *repo) CheckUserExisis(ctx context.Context, user model.User) (int64, error) {
	query := `SELECT id FROM "user" WHERE login=$1`
	var userID int64
	err := r.db.QueryRowContext(ctx, query, user.Login).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *repo) RegisterUser(ctx context.Context, user model.User) (int64, error) {
	var existingID int64
	query := `SELECT id FROM "user" WHERE login = $1`
	err := r.db.QueryRowContext(ctx, query, user.Login).Scan(&existingID)
	if err == nil {
		r.log.WithError(model.ErrLoginAlreadyTaken).Warning(model.ErrLoginAlreadyTaken.Error())
		return 0, model.ErrLoginAlreadyTaken
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Generate hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Insert new user
	insertQuery := `INSERT INTO "user" (login, password, balance) VALUES ($1, $2, $3) RETURNING id`
	var userID int64
	err = r.db.QueryRowContext(ctx, insertQuery, user.Login, hashedPassword, 0).Scan(&userID)
	if err != nil {
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
			return 0, model.ErrInvalidLoginAndPass
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		return 0, model.ErrInvalidLoginAndPass
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
		mbalance.Current = RoundToFiveDecimalPlaces(balance.Float64)
	}

	if withdrawn.Valid {
		mbalance.Withdrawn = RoundToFiveDecimalPlaces(withdrawn.Float64)
	}

	r.log.
		WithField("balance", mbalance.Current).
		Debug("GetBalance")

	return mbalance, nil
}
