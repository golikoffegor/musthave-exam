package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golikoffegor/musthave-exam/internal/mocks"
	"github.com/golikoffegor/musthave-exam/internal/model"
	"github.com/golikoffegor/musthave-exam/internal/repository"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestHandler_AddOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	logger := logrus.New()
	h := NewHandler(mockRepo, logger, nil)

	userID := int64(123)
	orderNumber := "12345678903"
	transaction := &model.Transaction{
		ID:     orderNumber,
		UserID: userID,
		Date:   time.Now(),
		Summ:   123456,
		Status: "NEW",
		Action: "Debit",
	}

	transactionNil := &model.Transaction{}
	tokenString, _ := h.BuildJWTString(userID)

	tests := []struct {
		name           string
		setupMock      func()
		authHeader     string
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Order added successfully",
			setupMock: func() {
				mockRepo.EXPECT().AddOrder(gomock.Any(), userID, orderNumber).Return(transaction, nil)
			},
			authHeader:     "Bearer " + tokenString,
			requestBody:    orderNumber,
			expectedStatus: http.StatusAccepted,
			expectedBody:   orderNumber,
		},
		{
			name: "Order already added by current user",
			setupMock: func() {
				mockRepo.EXPECT().AddOrder(gomock.Any(), userID, orderNumber).Return(transactionNil, nil)
			},
			authHeader:     "Bearer " + tokenString,
			requestBody:    orderNumber,
			expectedStatus: http.StatusOK,
			expectedBody:   orderNumber,
		},
		{
			name:           "Invalid order number",
			setupMock:      func() {},
			authHeader:     "Bearer " + tokenString,
			requestBody:    "123456",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   model.ErrInvalidOrderNumber.Error(),
		},
		{
			name:           "Server error while reading request body",
			setupMock:      func() {},
			authHeader:     "Bearer " + tokenString,
			requestBody:    "",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   model.ErrErrorRequestBody.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			var req *http.Request
			if tt.requestBody == "" {
				req = httptest.NewRequest("POST", "/add-order", &errorReader{})
			} else {
				req = httptest.NewRequest("POST", "/add-order", strings.NewReader(tt.requestBody))
			}
			req.Header.Set("Authorization", tt.authHeader)
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			h.AddOrderHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Contains(t, string(body), tt.expectedBody)
		})
	}
}

func TestHandler_GetOrdersHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := mocks.NewMockRepository(mockCtrl)
	mockLogger := logrus.New()
	h := NewHandler(mockRepo, mockLogger, nil)
	userID := int64(123)
	tokenString, _ := h.BuildJWTString(userID)
	cTime := time.Now()

	tests := []struct {
		name           string
		setupMock      func()
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			setupMock: func() {
				orders := []model.Transaction{
					{ID: "1", UserID: 123, Date: cTime, Summ: 100.0, Status: "NEW", Action: "Debit"},
				}
				mockRepo.EXPECT().GetOrders(gomock.Any(), int64(123), repository.Debit).Return(orders, nil)
			},
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusOK,
			expectedBody: func() string {
				orders := []model.Transaction{
					{ID: "1", UserID: 123, Date: cTime, Summ: 100.0, Status: "NEW", Action: "Debit"},
				}
				return mustMarshal(orders)
			}(),
		},
		{
			name: "Unauthorized",
			setupMock: func() {
			},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   model.ErrNotAuthorized.Error() + "\n",
		},
		{
			name: "InternalServerError",
			setupMock: func() {
				mockRepo.EXPECT().GetOrders(gomock.Any(), int64(123), repository.Debit).Return(nil, errors.New("internal error"))
			},
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
		{
			name: "NoContent",
			setupMock: func() {
				mockRepo.EXPECT().GetOrders(gomock.Any(), int64(123), repository.Debit).Return([]model.Transaction{}, nil)
			},
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("GET", "/orders", nil)
			req.Header.Set("Authorization", tt.authHeader)

			w := httptest.NewRecorder()

			h.GetOrdersHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, string(body))
		})
	}
}

func TestHandler_GetBalanceHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	logger := logrus.New()
	h := NewHandler(mockRepo, logger, nil)

	userID := int64(123)
	balance := model.Balance{
		Current:   1000.0,
		Withdrawn: 200.0,
	}
	balanceNil := model.Balance{}

	tokenString, _ := h.BuildJWTString(userID)

	tests := []struct {
		name           string
		setupMock      func()
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Get balance successfully",
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(gomock.Any(), userID).Return(balance, nil)
			},
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusOK,
			expectedBody:   string(mustMarshal(balance)),
		},
		{
			name:           "Unauthorized access",
			setupMock:      func() {},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   model.ErrNotAuthorized.Error(),
		},
		{
			name: "Internal server error",
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(gomock.Any(), userID).Return(balanceNil, errors.New("some error"))
			},
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("GET", "/get-balance", nil)
			req.Header.Set("Authorization", tt.authHeader)

			w := httptest.NewRecorder()
			h.GetBalanceHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Contains(t, string(body), tt.expectedBody)
		})
	}
}

func TestHandler_GetWithdrawalsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	logger := logrus.New()
	h := NewHandler(mockRepo, logger, nil)
	cTime := time.Now()
	userID := int64(123)
	withdrawals := []model.Withdraw{
		{
			ID:   "1",
			Date: cTime,
			Summ: 100.0,
		},
	}

	transactions := []model.Transaction{
		{
			ID:   "1",
			Summ: 100.0,
			Date: cTime,
		},
	}

	tokenString, _ := h.BuildJWTString(userID)

	tests := []struct {
		name           string
		setupMock      func()
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Get withdrawals successfully",
			setupMock: func() {
				mockRepo.EXPECT().GetOrders(gomock.Any(), userID, repository.Withdraw).Return(transactions, nil)
			},
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusOK,
			expectedBody:   string(mustMarshal(withdrawals)),
		},
		{
			name: "No withdrawals found",
			setupMock: func() {
				mockRepo.EXPECT().GetOrders(gomock.Any(), userID, repository.Withdraw).Return([]model.Transaction{}, nil)
			},
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "Unauthorized access",
			setupMock:      func() {},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   model.ErrNotAuthorized.Error(),
		},
		{
			name: "Internal server error",
			setupMock: func() {
				mockRepo.EXPECT().GetOrders(gomock.Any(), userID, repository.Withdraw).Return([]model.Transaction{}, errors.New("some error"))
			},
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("GET", "/get-withdrawals", nil)
			req.Header.Set("Authorization", tt.authHeader)

			w := httptest.NewRecorder()
			h.GetWithdrawalsHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Contains(t, string(body), tt.expectedBody)
		})
	}
}

func TestHandler_WithdrawHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	logger := logrus.New()
	h := NewHandler(mockRepo, logger, nil)

	userID := int64(123)
	tokenString, _ := h.BuildJWTString(userID)

	type errorModel struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
		Err   string  `json:"err"`
	}

	validWithdrawRequest := model.WithdrawRequest{
		Order: "12345678903",
		Sum:   100.0,
	}

	tests := []struct {
		name           string
		setupMock      func()
		authHeader     string
		requestBody    interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Withdraw successfully",
			setupMock: func() {
				mockRepo.EXPECT().Withdraw(gomock.Any(), userID, validWithdrawRequest.Order, validWithdrawRequest.Sum).Return(nil)
			},
			authHeader:     "Bearer " + tokenString,
			requestBody:    validWithdrawRequest,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		// {
		// 	name:           "Invalid request format",
		// 	setupMock:      func() {},
		// 	authHeader:     "Bearer " + tokenString,
		// 	requestBody:    "invalid request format",
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody:   "Invalid request format",
		// },
		{
			name:           "Invalid order number",
			setupMock:      func() {},
			authHeader:     "Bearer " + tokenString,
			requestBody:    model.WithdrawRequest{Order: "invalid_order", Sum: 100.0},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Invalid order number",
		},
		{
			name:           "Invalid json sum",
			setupMock:      func() {},
			authHeader:     "Bearer " + tokenString,
			requestBody:    model.WithdrawRequest{Order: "invalid_order"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "internal server error",
		},
		{
			name:       "Invalid json format",
			setupMock:  func() {},
			authHeader: "Bearer " + tokenString,
			requestBody: errorModel{
				Order: "12345678903",
				Err:   "err",
				Sum:   100.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "internal server error",
		},
		{
			name: "Insufficient funds",
			setupMock: func() {
				mockRepo.EXPECT().Withdraw(gomock.Any(), userID, validWithdrawRequest.Order, validWithdrawRequest.Sum).Return(errors.New(model.ErrIncFunds.Error()))
			},
			authHeader:     "Bearer " + tokenString,
			requestBody:    validWithdrawRequest,
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   model.ErrIncFunds.Error(),
		},
		{
			name:           "Unauthorized access",
			setupMock:      func() {},
			authHeader:     "",
			requestBody:    validWithdrawRequest,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   model.ErrNotAuthorized.Error(),
		},
		{
			name: "Internal server error",
			setupMock: func() {
				mockRepo.EXPECT().Withdraw(gomock.Any(), userID, validWithdrawRequest.Order, validWithdrawRequest.Sum).Return(errors.New("some error"))
			},
			authHeader:     "Bearer " + tokenString,
			requestBody:    validWithdrawRequest,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			var reqBody io.Reader
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					reqBody = bytes.NewBufferString(str)
				} else {
					bodyBytes, _ := json.Marshal(tt.requestBody)
					reqBody = bytes.NewBuffer(bodyBytes)
				}
			} else {
				reqBody = nil
			}

			req := httptest.NewRequest("POST", "/withdraw", reqBody)
			req.Header.Set("Authorization", tt.authHeader)

			w := httptest.NewRecorder()
			h.WithdrawHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Contains(t, string(body), tt.expectedBody)
		})
	}
}

func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
