package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/golikoffegor/musthave-exam/internal/mocks"
	"github.com/golikoffegor/musthave-exam/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func Test_repo_RegisterUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	logg := logrus.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	r := &repo{db: db, log: logg}

	ctx := context.Background()
	user := model.User{
		Login:    "testuser",
		Password: "password123",
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	assert.NotEmpty(t, hashedPassword)
	assert.NotEmpty(t, mockRepo)

	// Expectation for successful registration
	mock.ExpectQuery(`INSERT INTO "user" \(login, password, balance\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
		WithArgs(user.Login, sqlmock.AnyArg(), 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := r.RegisterUser(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// Expectation for login already taken
	mock.ExpectQuery(`INSERT INTO "user" \(login, password, balance\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
		WithArgs(user.Login, sqlmock.AnyArg(), 0).
		WillReturnError(sql.ErrNoRows)

	_, err = r.RegisterUser(ctx, user)
	assert.Equal(t, model.ErrLoginAlreadyTaken, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func Test_repo_LoginUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	logg := logrus.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	r := &repo{db: db, log: logg}

	ctx := context.Background()
	user := model.User{
		Login:    "testuser",
		Password: "password123",
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	assert.NotEmpty(t, hashedPassword)
	assert.NotEmpty(t, mockRepo)

	// Expectation for successful login
	mock.ExpectQuery(`SELECT id, password FROM "user" WHERE login=\$1`).
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).AddRow(1, hashedPassword))

	id, err := r.LoginUser(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// Expectation for invalid login password
	mock.ExpectQuery(`SELECT id, password FROM "user" WHERE login=\$1`).
		WithArgs(user.Login).
		WillReturnError(sql.ErrNoRows)

	_, err = r.LoginUser(ctx, user)
	assert.Equal(t, model.ErrInvalidLoginPass, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func Test_repo_GetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	logg := logrus.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	r := &repo{db: db, log: logg}
	assert.NotEmpty(t, mockRepo)

	ctx := context.Background()
	userID := int64(1)

	// Expectation for user found
	mock.ExpectQuery(`SELECT id, login, balance FROM "user" WHERE id=\$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "balance"}).AddRow(1, "testuser", 100.0))

	user, err := r.GetUser(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, 100.0, user.Balance)

	// Expectation for user not found
	mock.ExpectQuery(`SELECT id, login, balance FROM "user" WHERE id=\$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err = r.GetUser(ctx, userID)
	assert.Equal(t, model.ErrUserNotFound, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func Test_repo_GetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка при создании mock базы данных: %v", err)
	}
	defer db.Close()

	logg := logrus.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	r := &repo{db: db, log: logg}
	assert.NotEmpty(t, mockRepo)

	ctx := context.Background()
	userID := int64(1)

	// Expectation for balance found
	mock.ExpectQuery(regexp.QuoteMeta((`SELECT balance FROM "user" WHERE id = $1`))).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))

	mock.ExpectQuery(regexp.QuoteMeta((`SELECT SUM(summ) FROM transactions WHERE user_id = $1 AND action = 'Withdraw'`))).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(50.0))

	balance, err := r.GetBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, balance.Current)
	assert.Equal(t, 50.0, balance.Withdrawn)

	// Expectation for balance not found
	mock.ExpectQuery(regexp.QuoteMeta((`SELECT balance FROM "user" WHERE id = $1`))).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err = r.GetBalance(ctx, userID)
	assert.Error(t, err)

	// Expectation for failed to get withdrawn amount
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance FROM "user" WHERE id = $1`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT SUM(summ) FROM transactions WHERE user_id = $1 AND action = 'Withdraw'`)).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err = r.GetBalance(ctx, userID)
	assert.Error(t, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
