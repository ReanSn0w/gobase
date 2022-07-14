package messages

import (
	"context"
	"time"

	"github.com/ReanSn0w/gobase/pkg/utils"
	"github.com/ReanSn0w/mongo-monkey/wrap"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ChatCollection = "Chat"
)

// Функция обращается к БД для поиска чата с
func NewChat(clients ...primitive.ObjectID) (*Chat, error) {
	currentTime := time.Now()
	newChatID := primitive.NewObjectID()

	chat := &Chat{
		ID:       newChatID,
		SortTime: currentTime,
		Clients:  makeChatClients(currentTime, clients...),
	}

	return chat, chat.create()
}

// Отпавить сообщение в чат
func SendMessage(chatID, creatorID primitive.ObjectID, text string, media ...utils.Media) error {
	return utils.DB().Operation(func(ctx context.Context, w *wrap.Wrap) error {
		c := w.Collection(ChatCollection)

		_, err := c.UpdateOne(
			ctx,
			bson.D{
				{Key: "_id", Value: chatID},
				{Key: "clients.id", Value: creatorID},
			},
			bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "messages", Value: Message{
						Timestamp: time.Now(),
						Text:      text,
						Media:     media,
					}},
				}},
			},
		)

		return err
	})
}

func makeChatClients(currentTime time.Time, ids ...primitive.ObjectID) []Client {
	clients := make([]Client, 0, len(ids))

	for _, client := range ids {
		clients = append(clients, Client{
			ClientID:  client,
			Timestamp: currentTime,
		})
	}

	return clients
}
