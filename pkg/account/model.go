package account

import (
	"context"
	"time"

	"github.com/ReanSn0w/gobase/pkg/account/secure"
	"github.com/ReanSn0w/gobase/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	Сollection = "Account"
)

// Структура для описания аккаунта
type Account struct {
	ID     primitive.ObjectID `json:"id" bson:"_id"`    // Идентифкатор документа в коллекции
	Name   string             `json:"name" bson:"name"` // Человеко-читаемое имя пользователя в системе
	Secure secure.Secure      `json:"-" bson:"secure"`  // Секретная часть пользовательского профиля
}

// Создание нового аккаунта в системе
func CreateNewAccount(name string, group string, session secure.Session) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	s := secure.Secure{Sessions: []secure.Session{session}}

	res, err := utils.DB().Collection(Сollection).InsertOne(ctx, bson.D{
		{Key: "name", Value: name},
		{Key: "group", Value: group},
		{Key: "secure", Value: s},
	})
	if err != nil {
		return primitive.NilObjectID, err
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

// Удаление аккаунта пользователя из системы
func DeleteAccount(userID primitive.ObjectID) error {
	return utils.DB().DeleteObj(userID, Сollection)
}
