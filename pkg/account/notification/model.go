package notification

import (
	"time"

	"github.com/ReanSn0w/gobase/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	collection = "Account" // Данные уведомления хранятця в профиле пользователя
)

// Ключ для уведомления
type Key string

// идентификатор на который ссылается документ
type Target string

// Функция для создания уведомления
func CreateNotification(from primitive.ObjectID, target Target, key Key) *Notification {
	return &Notification{
		From:   from,
		Target: target,
		Key:    key,
	}
}

// Структура для добавления уведомления
type Notification struct {
	Index  int                `json:"index" bson:"index"`   // Поле индекса уведомления, появляется во время агрегации
	From   primitive.ObjectID `json:"from" bson:"from"`     // идентифицатор пользователя вызвавшего уведомление
	Target Target             `json:"target" bson:"target"` // ссылка на документ или deeplink
	Key    Key                `json:"key" bson:"key"`       // ключ уведомления
	Time   time.Time          `json:"time" bson:"time"`     // время создания уведомления
	Read   bool               `json:"read" bson:"read"`     // метка прочтения документа
}

// Метод отправки уведомлений пользователям
func (n *Notification) Send(profiles ...primitive.ObjectID) error {
	return utils.DB().UpdateSet(
		collection,
		bson.D{
			// тут надо проверить условие для обновления
			{Key: "_id", Value: bson.D{{Key: "$eq", Value: profiles}}},
		},
		bson.D{
			{Key: "$push", Value: bson.D{{Key: "notifications", Value: bson.D{
				{Key: "from", Value: n.From},
				{Key: "target", Value: n.Target},
				{Key: "key", Value: n.Key},
				{Key: "time", Value: time.Now()},
				{Key: "read", Value: false},
			}}}},
		},
	)
}
