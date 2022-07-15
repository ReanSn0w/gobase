package secure

import (
	"context"
	"errors"
	"time"

	"github.com/ReanSn0w/gobase/pkg/utils"
	"github.com/ReanSn0w/mongo-monkey/wrap"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNoUserIdClaim  = errors.New("в токене отсутствует user_id")
	ErrUnvalidUserID  = errors.New("user_id неверен или отсутствует")
	ErrNoGroupClaim   = errors.New("в токене отсутствует признак группу пользователя")
	ErrUnvalidGroup   = errors.New("признак группы неверен или отсутствует")
	ErrNoSessionClaim = errors.New("в токене отсутствует ключ сессии")
	ErrUnvalidSession = errors.New("клюх сессии неверен или отсутствует")
)

// Обновление токена пользователя для доступа к ресурсам
func RefreshUserToken(tokenString string) (string, error) {
	claims, err := utils.JWT().ParseUnverified(tokenString)
	if err != nil {
		return "", err
	}

	userID, group, sessionKey, err := parseClaims(claims)
	if err != nil {
		return "", err
	}

	err = checkUser(userID, group, sessionKey)
	if err != nil {
		return "", err
	}

	return CreateNewUserToken(userID, group, sessionKey)
}

func CreateNewUserToken(userID primitive.ObjectID, group string, sessionKey string) (string, error) {
	return utils.JWT().GenerateToken(jwt.MapClaims{
		"user_id":    userID.Hex(),
		"user_group": group,
		"session":    sessionKey,
		"iat":        time.Now().Add(time.Hour),
	})
}

func parseClaims(claims jwt.MapClaims) (primitive.ObjectID, string, string, error) {
	userIDClaim, ok := claims["user_id"]
	if !ok {
		return primitive.NilObjectID, "", "", ErrNoUserIdClaim
	}
	userIDString, ok := userIDClaim.(string)
	if !ok {
		return primitive.NilObjectID, "", "", ErrUnvalidUserID
	}
	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return primitive.NilObjectID, "", "", err
	}

	groupClaim, ok := claims["group"]
	if !ok {
		return primitive.NilObjectID, "", "", ErrNoGroupClaim
	}
	group, ok := groupClaim.(string)
	if !ok {
		return primitive.NilObjectID, "", "", ErrUnvalidGroup
	}

	sessionKeyClaim, ok := claims["session"]
	if !ok {
		return primitive.NilObjectID, "", "", ErrNoSessionClaim
	}
	session, ok := sessionKeyClaim.(string)
	if !ok {
		return primitive.NilObjectID, "", "", ErrUnvalidSession
	}

	return userID, group, session, nil
}

func checkUser(userID primitive.ObjectID, group string, sessionKey string) error {
	return utils.DB().Operation(func(ctx context.Context, w *wrap.Wrap) error {
		c := w.Collection(accountCollection)

		res := c.FindOne(
			ctx,
			bson.D{
				{Key: "_id", Value: userID},
				{Key: "secure.access", Value: group},
				{Key: "secure.sessions.key", Value: sessionKey},
			},
		)

		return res.Err()
	})
}
