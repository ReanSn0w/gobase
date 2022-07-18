package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

//Тип данных для возврата ошибки пользователю
type ErrorData struct {
	Code    int    // HTTP код ошибки
	Message string // Сообщение выводимое пользователю по интерфейсу Error
}

// Функция для отправки ошибки пользовател в виде стандартизированного объекта
func ResponseError(w http.ResponseWriter, code int, err error) {
	Response(w, code, ErrorData{Code: code, Message: err.Error()})
}

// Функция для отправки данных пользователю
func Response(w http.ResponseWriter, code int, obj interface{}) {
	w.WriteHeader(code)

	if obj != nil {
		encoder := json.NewEncoder(w)
		err := encoder.Encode(obj)
		if err != nil {
			log.Println(err)
		}
	}
}

// Мотнирует хранилище для статических файлов
func MountAssetsDirectory(r chi.Router) {
	r.Mount("/assets/", http.FileServer(http.Dir(".")))
}
