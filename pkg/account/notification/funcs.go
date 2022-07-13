package notification

import (
	"context"
	"fmt"

	"github.com/ReanSn0w/gobase/pkg/utils"
	"github.com/ReanSn0w/mongo-monkey/wrap"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Получение списка уведомлений для пользователя
func Get(profileID primitive.ObjectID, skip int, limit int) ([]Notification, error) {
	notifications := []Notification{}

	err := utils.DB().Operation(func(ctx context.Context, w *wrap.Wrap) error {
		c := w.Collection(collection)

		cur, err := c.Aggregate(ctx, mongo.Pipeline{
			{{Key: "$match", Value: bson.D{{Key: "_id", Value: profileID}}}},
			{{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$notifications"},
				{Key: "includeArrayIndex", Value: "index"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}}},
			{{Key: "$project", Value: bson.D{
				{Key: "_id", Value: "$index"},
				{Key: "from", Value: "$notifications.from"},
				{Key: "target", Value: "$notifications.target"},
				{Key: "key", Value: "$notifications.key"},
				{Key: "time", Value: "$notifications.time"},
				{Key: "read", Value: "$notifications.read"},
			}}},
			{{Key: "$sort", Value: bson.D{{Key: "time", Value: -1}}}},
			{{Key: "$skip", Value: skip}},
			{{Key: "$limit", Value: limit}},
		})
		if err != nil {
			return err
		}

		return cur.All(ctx, &notifications)
	})

	return notifications, err
}

// Установка метки о прочтении уведомления
func Read(profileID primitive.ObjectID, element int) error {
	return utils.DB().UpdateObj(profileID, collection, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: fmt.Sprintf("notifications.%v.read", element), Value: true},
		}},
	})
}
