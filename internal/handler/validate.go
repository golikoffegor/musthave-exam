package handler

import (
	"musthave-exam/internal/model"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

func (h *Handler) isValidAuth(_ http.ResponseWriter, r *http.Request) (int64, bool) {
	// session, _ := store.Get(r, "session")
	// UserID, ok := session.Values["user_id"].(int64)
	// if !ok {
	// 	return 0, false
	// }
	// return UserID, true

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, false
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, false
	}

	tokenString := parts[1]
	claims := &model.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		return -1, false
	}

	if !token.Valid {
		return -1, false
	}

	return claims.UserID, true
}

func (h *Handler) setAuth(w http.ResponseWriter, _ *http.Request, userID int64) error {
	// session, _ := store.Get(r, "session")
	// session.Values["user_id"] = userID
	// session.Save(r, w)

	token, err := h.BuildJWTString(userID)
	if err != nil {
		return err
	}
	// token := h.generateAuthToken(userID)
	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "Authorization",
	// 	Value:    token,
	// 	Path:     "/",
	// 	Expires:  time.Now().Add(24 * time.Hour),
	// 	HttpOnly: true,
	// })
	w.Header().Set("Authorization", "Bearer "+token)
	return nil
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (h *Handler) BuildJWTString(userID int64) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, model.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (h *Handler) isValidOrderNumber(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}
		if double {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		double = !double
	}
	return sum%10 == 0
}
