package secure

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	userIDCtxKey      = &ctxKeyUID{}
	userGroupCtxKey   = &ctxKeyGID{}
	userSessionCtxKey = &ctxKeySID{}
)

type ctxKeyUID struct{}
type ctxKeyGID struct{}
type ctxKeySID struct{}

// Обновление контекста для запроса
func buildusercontext(ctx context.Context, userID primitive.ObjectID, userGroup string, userSession string) context.Context {
	ctx = context.WithValue(ctx, userIDCtxKey, userID)
	ctx = context.WithValue(ctx, userGroupCtxKey, userGroup)
	ctx = context.WithValue(ctx, userSessionCtxKey, userSession)
	return ctx
}

// Получениеи идентификатора пользователя из контекста
func UserIDFromContext(ctx context.Context) primitive.ObjectID {
	return ctx.Value(userIDCtxKey).(primitive.ObjectID)
}

// Получение группы пользователя из контекста
func UserGroupFromContext(ctx context.Context) string {
	return ctx.Value(userGroupCtxKey).(string)
}
