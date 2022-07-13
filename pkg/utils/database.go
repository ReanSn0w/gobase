package utils

import (
	"github.com/ReanSn0w/mongo-monkey/wrap"
)

var (
	db *wrap.Wrap

	systemCollection = "system"
)

// Функция эмулирует синглтон для доступа к обёртке БД
func DB() *wrap.Wrap {
	return db
}

// Функция настроки базы данных на основе окружения
func ConfigureDB() error {
	wrap, err := wrap.CreateWrapFromEnv()
	if err != nil {
		return err
	}

	db = wrap
	return wrap.Connect()
}
