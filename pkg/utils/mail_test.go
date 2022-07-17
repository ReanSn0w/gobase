package utils_test

import (
	"testing"

	"github.com/ReanSn0w/gobase/pkg/utils"
)

func Test_SendPlainMail(t *testing.T) {
	mail := utils.NewPlainMail(
		"Дмитрий Папков",
		"papkovda@me.com",
		"Test mail",
		"Текст сообщения",
	)

	sendMainTestBase(mail, t)
}

func Test_SendHTMLMail(t *testing.T) {
	mail := utils.NewHtmlMail(
		"Дмитрий Папков",
		"papkovda@me.com",
		"Тест отправки сообщения",
		[]byte("<html><body><h1>hello, world</h1></body></html>"),
	)

	sendMainTestBase(mail, t)
}

func sendMainTestBase(mail utils.Email, t *testing.T) {
	utils.Mailer.SetConfiguration(
		"smtp.yandex.ru",
		"465",
		"Дмитрий Папков",
		"reansnow@yandex.ru",
		"reansnow",
		"omunigqoqwvqvpqc",
	)

	err := utils.Mailer.Send(mail)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
