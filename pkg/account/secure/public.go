package secure

import (
	"errors"

	"github.com/ReanSn0w/gobase/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	accountCollection = "Account"

	ErrSessionUnvalid = errors.New("сессия пользователя не валидна")
)

// Функция создает новую сессию для пользователя с установленным названием
//
// Название может быть UserAgent'ом браузера или названием телефона
func CreateSession(name string) Session {
	return Session{
		Name: name,
		Key:  utils.GenerateRandomString(24, true, true, false),
	}
}

// Фнкция добавляет новую сессию пользователя к профилю
func AppendSession(userID primitive.ObjectID, session Session) error {
	return utils.DB().UpdateObj(userID, accountCollection, bson.D{
		{
			Key: "$push",
			Value: bson.D{
				{Key: "secure.sessions", Value: session},
			},
		},
	})
}

// Удаление сессий пользователя
//
// Удаляет все сессии пользователя, пользователь в данном случае должен быть разлогинен
// если требуется оставить пользователя залогиненным, следует создать новую сессию для него и выдать новый токен
func RemoveAllSessions(userID primitive.ObjectID) error {
	return utils.DB().UpdateObj(userID, accountCollection, bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "secure.sessions", Value: []Session{}},
			},
		},
	})
}

// Функция для проверки сесси пользователя
//
// В нормально случае возвращает пустой интерфейс, во всех остальных следует разлогинить пользователя
func CheckSession(userID primitive.ObjectID, sessionKey string) error {
	count, err := utils.DB().CountElements(accountCollection, bson.D{
		{Key: "_id", Value: userID},
		{Key: "secure.session.key", Value: sessionKey},
	})

	if err != nil || count != 1 {
		return ErrSessionUnvalid
	}

	return nil
}
