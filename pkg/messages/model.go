package messages

import (
	"time"

	"github.com/ReanSn0w/gobase/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Структура описывает сообшения в чате
type Chat struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`            // Уникальный идентификатор чата
	SortTime time.Time          `json:"-" bson:"sort_time"`       // Время для сортировки сообщений
	Clients  []Client           `json:"clients" bson:"clients"`   // Учасники чата
	Messages []Message          `json:"messages" bson:"messages"` // Сообщения чата
}

func (c *Chat) create() error {
	res, err := utils.DB().CreateObj(c)
	if err != nil {
		return err
	}

	c.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// Структура описывает пользователя чата
type Client struct {
	ClientID  primitive.ObjectID `json:"id" bson:"id"`               // Идентификатор пользователя чата
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"` // Метка последлего открытия чата, для определения наличия новых сообщений
}

// Структура описывает сообщение в чате
type Message struct {
	Timestamp time.Time     // Сообщение в чате
	Text      string        // Текст сообщения
	Media     []utils.Media // Ссылки на мультимедийный контент
}
