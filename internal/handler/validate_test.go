package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/golikoffegor/musthave-exam/internal/mocks"
	"github.com/golikoffegor/musthave-exam/internal/model"
	"github.com/golikoffegor/musthave-exam/internal/settings"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHandler_isValidAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	logger := logrus.New()
	parsed := settings.Parse()
	h := NewHandler(mockRepo, logger, parsed)

	claims := &model.Claims{
		UserID: 123,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()

	userID, ok := h.isValidAuth(rr, req)

	assert.True(t, ok, "expected true, got false")
	assert.Equal(t, int64(123), userID, "expected userID to be 123, got %d", userID)

}
