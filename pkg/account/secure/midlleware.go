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
	ErrUnvalidToken  = errors.New("ошибка, токен не валидирован или устарел")
	ErrRequestLocked = errors.New("запрос завлокирован так так у пользователя недостаточно полномочий на его выполнение")
)

// Метод для проверки доступа к действию
//
// В случае если у пользователя достаточно полномочий, его запрос перейдет дальше,
// однако если полномочий недостаточно, запрос будет завершен с кодом 423
func CheckPrivilegeMiddleware(privileges ...utils.PrivilegeType) func(http.Handler) http.Handler {
	return CheckModulePrivilegeMiddleware("main", privileges...)
}

// Метод для проверки доступа к действию по модулю
//
// В случае если у пользователя достаточно полномочий, его запрос перейдет дальше,
// однако если полномочий недостаточно, запрос будет завершен с кодом 423
func CheckModulePrivilegeMiddleware(module string, privileges ...utils.PrivilegeType) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := UserIDFromContext(ctx)
			userGroup := UserGroupFromContext(ctx)

			if utils.Privileges().Check(userID, userGroup, module, privileges...) {
				h.ServeHTTP(w, r)
			} else {
				utils.ResponseError(w, http.StatusLocked, ErrRequestLocked)
			}
		})
	}
}

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
