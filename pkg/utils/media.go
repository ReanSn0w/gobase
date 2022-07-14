package utils

const (
	MediaImage   MediaType = "image"   // ссылка на изображение
	MediaVideo   MediaType = "video"   // ссылка на видео
	MediaYoutube MediaType = "youtube" // ссылка на ролик на Youtube
	MediaLink    MediaType = "link"    // Ссылка на сайт
)

type MediaType string

// Тип представляет из себя ссылку на медиа контент
type Media string

// Метод визвращает тип документа на основе его ссылки
func (m Media) Type() MediaType {
	return MediaLink
}
