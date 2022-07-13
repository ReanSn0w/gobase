package classic

import (
	"context"
	"errors"
	"time"

	"github.com/ReanSn0w/gobase/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	accountCollection = "Account"

	ErrEmailUnavaliable = errors.New("данный email уже используется")
)

// Данные об авторизации пользователя стандартными средствами (Email/Пароль)
type ClassicAuth struct {
	ID    primitive.ObjectID `bson:"_id"`   // Идентификатор пользователя в системе
	Group string             `bson:"group"` // группа к который пренадлежит пользователь
	Email string             `bson:"email"` // email пользователя
	Hash  []byte             `bson:"hash"`  // пароль пользователя
}

func (ca *ClassicAuth) Validate(password string) bool {
	err := bcrypt.CompareHashAndPassword(ca.Hash, []byte(password))

	if err != nil {
		return true
	} else {
		return false
	}
}

// сохранение данных об авторизации пользователя
func saveUserCredentials(userID primitive.ObjectID, email string, hash []byte) error {
	return utils.DB().UpdateObj(userID, accountCollection, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "secure.auth.classic.email", Value: email},
			{Key: "secure.auth.classic.hash", Value: hash},
		}},
	})
}

// Загрузка данных для авторизации пользователя
func loadUserCredentialsByEmail(email string) (*ClassicAuth, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	cur, err := utils.DB().Collection(accountCollection).Aggregate(ctx, mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "secure.auth.classic.email", Value: email},
			}},
		},
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: "$_id"},
				{Key: "group", Value: "$secure.access"},
				{Key: "email", Value: "$secure.auth.classic.email"},
				{Key: "hash", Value: "$secure.auth.classic.hash"},
			}},
		},
	})
	if err != nil {
		return nil, err
	}

	auth := &ClassicAuth{}
	for cur.Next(ctx) {
		err := cur.Decode(auth)
		if err != nil {
			return nil, err
		} else {
			break
		}
	}

	return auth, nil
}

func emailAvaliable(email string) error {
	count, err := utils.DB().CountElements(accountCollection, bson.D{{Key: "secure.auth.classic.email", Value: email}})
	if err != nil {
		return err
	}

	if count != 0 {
		return ErrEmailUnavaliable
	}

	return nil
}
