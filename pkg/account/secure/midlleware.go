package secure

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/ReanSn0w/gobase/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	accessTokenCookie = "access_token"
	accessTokenHeader = "Authorization"
)

var (
	userIDCtxKey      = &ctxKey{}
	userGroupCtxKey   = &ctxKey{}
	userSessionCtxKey = &ctxKey{}

	ErrUnvalidToken = errors.New("ошибка, токен не валидирован или устарел")
)

type ctxKey struct{}

// Middleware для авторизации пользователя для сайта
//
// Проверит наличие токена в cookie запроса
// В случае если токена нет запишет в контекст значения для гостя и продолжит выполнение
// В случае если токен не валиден проверит сессию
//    если сессия валидна обновит токен и продолжит выполнение запроса с данными пользователя
//    если нет удалит токен, запишет значения для гостя и продолжит выполнение
// В нормальном состоянии запишет UID, группу и клыч сессии в контекст и продожит выполнение
func SiteAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(accessTokenCookie)
		if err != nil {
			// Не удалось получить токен пользователя
			log.Println(err)
			ctx := userguestvalues(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx, err := checktoken(r.Context(), tokenCookie.Value)
		if err != nil {
			userID, group, session, err := unverifiedchecktoken(tokenCookie.Value)
			if err != nil {
				// Не удалось достать значения из строки токена
				log.Println(err)
				removeTokenCookie(w)
				ctx := userguestvalues(r.Context())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// сессия пользователя проверена, производится выпуск нового токена
			tokenString, err := CreateNewUserToken(userID, group, session)
			if err != nil {
				// Не удалось достать значения из строки токена
				log.Println(err)
				removeTokenCookie(w)
				ctx := userguestvalues(r.Context())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			WriteTokenCookie(w, tokenString)
			ctx = buildusercontext(ctx, userID, group, session)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware проверки пользователя для API
//
// В случае отсутствия токена вернет 401 код и завешит выполнение запроса
// В случае если токен нужно обновить вернет 412 код и завершит выполнение запроса
// В нормальном состоянии запишет UID, группу и клыч сессии в контекст и продожит выполнение
func APIAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get(accessTokenHeader)
		ctx, err := checktoken(r.Context(), tokenString)
		if err != nil {
			log.Println(err)

			_, _, _, err = unverifiedchecktoken(tokenString)
			if err != nil {
				log.Println(err)
				utils.ResponseError(w, 401, err)
				return
			}

			utils.ResponseError(w, 412, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WriteTokenCookie(w http.ResponseWriter, tokenString string) {
	cookie := &http.Cookie{
		Name:   accessTokenCookie,
		Value:  tokenString,
		MaxAge: 172800,
	}

	http.SetCookie(w, cookie)
}

func removeTokenCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   accessTokenCookie,
		Value:  "",
		MaxAge: -1,
	}

	http.SetCookie(w, cookie)
}

// Проверяет токен на валидность, в случае если токен валиден
// прозиводится запись сначений в контекст и запрос отправляется дальше
// на обработку
func checktoken(ctx context.Context, tokenString string) (context.Context, error) {
	claims, err := utils.JWT().Parse(tokenString)
	if err != nil {
		return ctx, ErrUnvalidToken
	}

	userID, userGroup, userSession, err := parseClaims(claims)
	if err != nil {
		return ctx, err
	}

	return buildusercontext(ctx, userID, userGroup, userSession), nil
}

// Установка значений гостя в контекст
func userguestvalues(ctx context.Context) context.Context {
	return buildusercontext(ctx, primitive.NilObjectID, "guest", "")
}

// Обновление контекста для запроса
func buildusercontext(ctx context.Context, userID primitive.ObjectID, userGroup string, userSession string) context.Context {
	ctx = context.WithValue(ctx, userIDCtxKey, userID)
	ctx = context.WithValue(ctx, userGroupCtxKey, userGroup)
	ctx = context.WithValue(ctx, userSessionCtxKey, userSession)
	return ctx
}

// Проверка возможности обновления токена
func unverifiedchecktoken(tokenString string) (primitive.ObjectID, string, string, error) {
	// Проверить на возможность обновления токена
	claims, err := utils.JWT().ParseUnverified(tokenString)
	if err != nil {
		return primitive.NilObjectID, "", "", err
	}

	userID, group, session, err := parseClaims(claims)
	if err != nil {
		return primitive.NilObjectID, "", "", err
	}

	err = checkUser(userID, group, session)
	return primitive.NilObjectID, "", "", err
}
