package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golikoffegor/musthave-exam/internal/mocks"
	"github.com/golikoffegor/musthave-exam/internal/model"
	"github.com/golikoffegor/musthave-exam/internal/settings"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_handler_RegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	logger := logrus.New()
	parsed := &settings.InitedFlags{}
	h := NewHandler(mockRepo, logger, parsed)

	user := model.User{
		ID:       1,
		Login:    "testuser",
		Password: "testpassword",
	}

	//регаем юзера
	logger.Info("Начало регистрации пользователя")
	body, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/register", bytes.NewReader(body))

	assert.NoError(t, err)

	mockRepo.EXPECT().RegisterUser(gomock.Any(), gomock.Eq(user)).Return(user.ID, nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.RegisterHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// assert.Contains(t, logger, "userID")

	//пустой запрос
	logger.Info("Начало пустой запрос")
	req, err = http.NewRequest("POST", "/register", bytes.NewReader([]byte("")))

	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	//ошибка json
	logger.Info("Начало ошибка json")
	req, err = http.NewRequest("POST", "/register", bytes.NewReader([]byte("{")))
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var uID int64 = 0
	// пользователь существует
	logger.Info("Начало если пользователь существует")
	mockRepo.EXPECT().RegisterUser(gomock.Any(), gomock.Eq(user)).Return(uID, model.ErrLoginAlreadyTaken)

	req, err = http.NewRequest("POST", "/register", bytes.NewReader(body))
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	// ошибкм сервера
	logger.Info("Начало если ошибка сервера")
	mockRepo.EXPECT().RegisterUser(gomock.Any(), gomock.Eq(user)).Return(uID, errors.New("some internal error"))

	req, err = http.NewRequest("POST", "/register", bytes.NewReader(body))
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func Test_handler_LoginUserHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	logger := logrus.New()
	parsed := &settings.InitedFlags{}
	h := NewHandler(mockRepo, logger, parsed)

	credentials := model.User{
		Login:    "testuser",
		Password: "testpassword",
	}

	logger.Info("Начало логин пользователя")
	body, _ := json.Marshal(credentials)
	req, err := http.NewRequest("POST", "/login", bytes.NewReader(body))
	assert.NoError(t, err)

	mockRepo.EXPECT().LoginUser(gomock.Any(), gomock.Eq(credentials)).Return(int64(1), nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.LoginUserHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	logger.Info("Начало ошибка при пустом теле запроса")
	req, err = http.NewRequest("POST", "/login", bytes.NewReader([]byte("")))
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	logger.Info("Начало ошибка при неправильном json")
	req, err = http.NewRequest("POST", "/login", bytes.NewReader([]byte("{")))
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	logger.Info("Начало ошибка неверный логин/пароль")
	mockRepo.EXPECT().LoginUser(gomock.Any(), gomock.Eq(credentials)).Return(int64(0), errors.New(model.ErrLoginAlreadyTaken.Error()))

	req, err = http.NewRequest("POST", "/login", bytes.NewReader(body))
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	logger.Info("Начало ошибка сервера")
	mockRepo.EXPECT().LoginUser(gomock.Any(), gomock.Eq(credentials)).Return(int64(0), errors.New("some internal error"))

	req, err = http.NewRequest("POST", "/login", bytes.NewReader(body))
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func Test_handler_GetUserHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	logger := logrus.New()
	parsed := &settings.InitedFlags{}
	h := NewHandler(mockRepo, logger, parsed)

	userID := int64(123)
	user := model.User{
		ID:       userID,
		Login:    "testuser",
		Password: "testpassword",
	}

	tokenString, _ := h.BuildJWTString(userID)

	mockRepo.EXPECT().GetUser(gomock.Any(), gomock.Eq(userID)).Return(user, nil)

	req, err := http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.GetUserHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var responseUser model.User
	err = json.NewDecoder(rr.Body).Decode(&responseUser)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, user.ID, responseUser.ID)
	assert.Equal(t, user.Login, responseUser.Login)
	assert.Equal(t, user.Password, responseUser.Password)

	req, err = http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer invalid_token")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	req, err = http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	mockRepo.EXPECT().GetUser(gomock.Any(), gomock.Eq(userID)).Return(model.User{}, model.ErrUserNotFound)

	req, err = http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
