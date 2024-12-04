package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/golikoffegor/musthave-exam/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_repo_AddOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	logg := logrus.New()
	r := &repo{db: mockDB, log: logg}

	ctx := context.Background()
	userID := int64(1)
	orderNumber := "order123"

	mock.ExpectQuery(`SELECT id, user_id FROM transactions WHERE id = \$1 LIMIT 1`).
		WithArgs(orderNumber).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))

	_, err = r.AddOrder(ctx, userID, orderNumber)
	assert.Error(t, err)

	//TODO: декомпозировать
	// mock.ExpectQuery(`SELECT id, user_id FROM transactions WHERE id = \$1 LIMIT 1`).
	// 	WithArgs(orderNumber).
	// 	WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))

	// mock.ExpectQuery(`INSERT INTO transactions (id, user_id, summ, status, action) VALUES (\$1, \$2, \$3, 'NEW', 'Debit') RETURNING id, user_id, summ, status, action, date`).
	// 	WithArgs(orderNumber, userID, 0).
	// 	WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "summ", "status", "action", "date"}).
	// 		AddRow(orderNumber, userID, 0, "NEW", "Debit", time.Now()))

	// transaction, err := r.AddOrder(ctx, userID, orderNumber)
	// if err != nil {
	// 	t.Errorf("Unexpected error: %v", err)
	// }
	// if transaction == nil {
	// 	t.Error("Expected transaction to be returned, got nil")
	// } else {
	// 	t.Logf("Transaction added successfully: %v", transaction)
	// }

	// Проверяем ошибки
	// assert.NoError(t, err)
	// assert.NotNil(t, transaction)
	// assert.Equal(t, userID, transaction.UserID)
	// assert.Equal(t, orderNumber, transaction.ID)
	// assert.Equal(t, float64(0), transaction.Summ)
	// assert.Equal(t, "NEW", transaction.Status)
	// assert.Equal(t, "Debit", transaction.Action)

	// Case 2: Order exists for a different user
	mock.ExpectQuery(`SELECT id, user_id FROM transactions WHERE id = \$1 LIMIT 1`).
		WithArgs(orderNumber).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).
			AddRow(1, userID+1))

	transaction, err := r.AddOrder(ctx, userID, orderNumber)
	if err != model.ErrAddExistsOrder {
		t.Errorf("Expected error ErrAddExistsOrder, got: %v", err)
	}
	if transaction != nil {
		t.Error("Expected transaction to be nil, got:", transaction)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func Test_repo_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	logg := logrus.New()
	r := &repo{db: mockDB, log: logg}

	ctx := context.Background()
	userID := int64(1)
	orderNumber := "123456"
	sum := 50.0
	initialBalance := 100.0
	remainingBalance := initialBalance - sum

	logg.Info("remainingBalance ", remainingBalance)

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT balance FROM "user" WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(initialBalance))

	mock.ExpectExec(`UPDATE "user" SET balance = balance - \$1 WHERE id = \$2`).
		WithArgs(sum, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO transactions (id, user_id, summ, status, action) VALUES ($1, $2, $3, 'NEW', 'Withdraw')`)).
		WithArgs(orderNumber, userID, sum).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = r.Withdraw(ctx, userID, orderNumber, sum)

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не выполнены: %s", err)
	}
}

func Test_repo_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	logg := logrus.New()
	r := &repo{db: mockDB, log: logg}

	ctx := context.Background()
	userID := int64(1)
	act := "Withdraw"
	now := time.Now().Truncate(time.Second) // Убираем миллисекунды

	mock.ExpectQuery(`SELECT id, status, action, date, summ FROM transactions WHERE user_id = \$1 AND action = \$2 ORDER BY date`).
		WithArgs(userID, act).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "action", "date", "summ"}).
			AddRow("1", "NEW", "Withdraw", now, 100.0).
			AddRow("2", "COMPLETED", "Withdraw", now, 50.0))

	orders, err := r.GetOrders(ctx, userID, act)

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, "1", orders[0].ID)
	assert.Equal(t, "NEW", orders[0].Status)
	assert.Equal(t, "Withdraw", orders[0].Action)
	assert.Equal(t, now, orders[0].Date)
	assert.Equal(t, 100.0, orders[0].Summ)
	assert.Equal(t, "2", orders[1].ID)
	assert.Equal(t, "COMPLETED", orders[1].Status)
	assert.Equal(t, "Withdraw", orders[1].Action)
	assert.Equal(t, now, orders[1].Date)
	assert.Equal(t, 50.0, orders[1].Summ)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не выполнены: %s", err)
	}
}

func Test_repo_UpdateTransactionStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	logg := logrus.New()
	r := &repo{db: mockDB, log: logg}

	ctx := context.Background()
	transactionID := "12345"
	status := "COMPLETED"

	mock.ExpectExec(`UPDATE transactions SET status = \$1 WHERE id = \$2`).
		WithArgs(status, transactionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = r.UpdateTransactionStatus(ctx, transactionID, status)

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не выполнены: %s", err)
	}
}

func Test_repo_UpdateTransactionStatusAndAccrual(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	logg := logrus.New()
	r := &repo{db: mockDB, log: logg}

	ctx := context.Background()
	transactionID := "12345"
	status := "PROCESSED"
	accrual := 100.0
	userID := int64(1)

	mock.ExpectBegin()

	mock.ExpectExec(`UPDATE transactions SET status = \$1, summ = \$2 WHERE id = \$3`).
		WithArgs(status, accrual, transactionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(`SELECT user_id FROM transactions WHERE id = \$1`).
		WithArgs(transactionID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(userID))

	mock.ExpectExec(`UPDATE "user" SET balance = balance \+ \$1 WHERE id = \$2`).
		WithArgs(accrual, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = r.UpdateTransactionStatusAndAccrual(ctx, transactionID, status, accrual)

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не выполнены: %s", err)
	}
}

func Test_repo_GetNewTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	logg := logrus.New()
	r := &repo{db: mockDB, log: logg}

	ctx := context.Background()
	currentTime := time.Now()

	mock.ExpectQuery(`SELECT id, user_id, summ, date, status, action FROM transactions WHERE action = 'Debit' AND status = 'NEW'`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "summ", "date", "status", "action"}).
			AddRow("1", 1, 100.0, currentTime, "NEW", "Debit").
			AddRow("2", 2, 200.0, currentTime, "NEW", "Debit"))

	transactions, err := r.GetNewTransactions(ctx)

	assert.NoError(t, err)
	assert.Len(t, transactions, 2)
	assert.Equal(t, "1", transactions[0].ID)
	assert.Equal(t, int64(1), transactions[0].UserID)
	assert.Equal(t, 100.0, transactions[0].Summ)
	assert.Equal(t, currentTime, transactions[0].Date)
	assert.Equal(t, "NEW", transactions[0].Status)
	assert.Equal(t, "Debit", transactions[0].Action)

	assert.Equal(t, "2", transactions[1].ID)
	assert.Equal(t, int64(2), transactions[1].UserID)
	assert.Equal(t, 200.0, transactions[1].Summ)
	assert.Equal(t, currentTime, transactions[1].Date)
	assert.Equal(t, "NEW", transactions[1].Status)
	assert.Equal(t, "Debit", transactions[1].Action)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не выполнены: %s", err)
	}
}
