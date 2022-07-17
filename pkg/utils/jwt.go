package utils

import (
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt"
)

var (
	tokenizer *jwtUtility = &jwtUtility{jwt: jwtauth.New("HS256", []byte(Salt()), nil)}

	ErrNoToken      = errors.New("не удалось извлечь JWT токен из запроса")
	ErrUnvalidToken = errors.New("токен пользователя не является валидным")
)

func JWT() *jwtUtility {
	return tokenizer
}

type jwtUtility struct {
	jwt *jwtauth.JWTAuth
}

// Метод для создания Middleware, аутентификации пользователя
//
// action - действие для модификации контекста на основе claims из запроса или ошибки
func (utility *jwtUtility) Authentificator(action func(*http.Request, jwt.MapClaims, error) *http.Request) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Получение токена в контекст из запроса
		verifier := jwtauth.Verifier(utility.jwt)
		next = verifier(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				r = action(r, claims, ErrNoToken)
			} else {
				// Действие в случае успешного получение токена пользователя из запроса
				r = action(r, claims, nil)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Метод для генерации нового токена авторизации пользователя
func (utility *jwtUtility) GenerateToken(claims jwt.MapClaims) (string, error) {
	_, token, err := utility.jwt.Encode(claims)
	return token, err
}

// Получение массива claims из токена без валидации токена
func (utility *jwtUtility) ParseUnverified(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	return claims, err
}

// Получение массива claims из токена
func (utility *jwtUtility) Parse(tokenString string) (jwt.MapClaims, error) {
	token, err := utility.jwt.Decode(tokenString)
	if err != nil {
		return nil, err
	}

	return token.PrivateClaims(), nil
}
