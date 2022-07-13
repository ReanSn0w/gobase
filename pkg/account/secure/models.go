package secure

import "time"

// Структура для описания сессии пользователя
type Session struct {
	Name       string    `bson:"name"`        // Название сессии (является произвольным полем, однако корректно его использовать для описания сущьности с который был произведен вход)
	Key        string    `bson:"key"`         // Строка сохраняемая в токен пользователю, используется для обновления токена
	LastUpdate time.Time `bson:"last_update"` // Время последнего обращения к сессии
}

// Структура для сохранения секретной информации о пользователе
type Secure struct {
	Access   string                 `json:"-" bson:"access"` // Идентификатор группы
	AuthData map[string]interface{} `bson:"auth"`            // Данные для авторизации пользователя
	Sessions []Session              `bson:"sessions"`        // Данные о сессиях пользователя
}
