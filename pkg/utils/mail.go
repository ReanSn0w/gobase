package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
)

var (
	Mailer *mailer = &mailer{}
)

// Структура для работы с Email рассылками
type mailer struct {
	config *mailerconfig
}

func (m *mailer) Send(message Email) error {
	if m.config == nil {
		return errors.New("отсутствует конфугурация для smtp сервера")
	}
	return m.sendMail(message)
}

func (m *mailer) addresses(message Email) (from, to mail.Address) {
	from = mail.Address{Name: m.config.Name, Address: m.config.Email}
	to = mail.Address{Name: message.Recipient(), Address: message.RecipientEmail()}
	return
}

func (m *mailer) message(from, to mail.Address, message Email) []byte {
	buffer := new(bytes.Buffer)

	buffer.WriteString(fmt.Sprintf("From: %s\r\n", from.String()))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", to.String()))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n\r\n", message.MessageSubject()))
	buffer.WriteString("MIME-Version: 1.0\r\n")
	buffer.WriteString(fmt.Sprintf("Content-Type: %s; charset=\"UTF-8\"\r\n", message.MessageContent()))
	buffer.Write(message.MessageContent())

	return buffer.Bytes()
}

func (m *mailer) client() (*smtp.Client, error) {
	servername := m.config.Host + ":" + m.config.Port
	host, _, _ := net.SplitHostPort(servername)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         m.config.Host,
	}

	// Вызов tcp соедиения
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return nil, err
	}

	return smtp.NewClient(conn, host)
}

func (m *mailer) sendMail(mail Email) error {
	from, to := m.addresses(mail)
	auth := smtp.PlainAuth("", m.config.Login, m.config.Password, m.config.Host)
	client, err := m.client()
	if err != nil {
		return err
	}

	err = client.Auth(auth)
	if err != nil {
		return err
	}

	err = client.Mail(from.Address)
	if err != nil {
		return err
	}

	err = client.Rcpt(to.Address)
	if err != nil {
		return err
	}

	wr, err := client.Data()
	defer func() {
		err := wr.Close()

		if err != nil {
			log.Println(err)
		}
	}()
	if err != nil {
		return err
	}

	_, _ = wr.Write(m.message(from, to, mail))
	// if err != nil {
	// 	return err
	// }

	return client.Quit()
}

type mailerconfig struct {
	Host     string `bson:"host"`
	Port     string `bson:"port"`
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Login    string `bson:"login"`
	Password string `bson:"password"`
}

// Загрузка конфигурации для отправки сообщений
func (m *mailer) Load() error {
	if m.config != nil {
		return nil
	}

	conf := mailerconfig{}
	err := Configuration.Load("mailer", &conf)
	if err != nil {
		return err
	}

	m.config = &conf
	return nil
}

// Установка произвольной конфигурации для mail клиента
func (m *mailer) SetConfiguration(host, port, name, email, login, password string) {
	m.config = &mailerconfig{
		Host:     host,
		Port:     port,
		Name:     name,
		Email:    email,
		Login:    login,
		Password: password,
	}
}

type Email interface {
	Recipient() string      // Имя получателя сообения
	RecipientEmail() string // Email получателя сообщения
	MessageSubject() string // Тема письма
	MessageType() string    // ContentType письма
	MessageContent() []byte // Сообщение
}

// Создает новое сообзение с HTML разметкой
func NewHtmlMail(name string, email string, subject string, message []byte) Email {
	return &HTMLEmail{
		Name:    name,
		Email:   email,
		Subject: subject,
		Message: message,
	}
}

// Структура для отправки сообщения с HTML разметкой
type HTMLEmail struct {
	Name    string
	Email   string
	Subject string
	Message []byte
}

// Имя получателя
func (m *HTMLEmail) Recipient() string {
	return m.Name
}

// Email получателя
func (m *HTMLEmail) RecipientEmail() string {
	return m.Email
}

// Тема сообщения
func (m *HTMLEmail) MessageSubject() string {
	return m.Subject
}

// ContentType
func (m *HTMLEmail) MessageType() string {
	return "text/html"
}

// Тело сообщения
func (m *HTMLEmail) MessageContent() []byte {
	return m.Message
}

// Создает новое простое сообщение
func NewPlainMail(name, email, sub, message string) Email {
	return &PlainMail{
		Name:    name,
		Email:   email,
		Subject: sub,
		Message: message,
	}
}

// Структура для отправки простого письма
type PlainMail struct {
	Name    string
	Email   string
	Subject string
	Message string
}

// Имя адресата сообщения
func (pm *PlainMail) Recipient() string {
	return pm.Name
}

// Email адресата сообщения
func (pm *PlainMail) RecipientEmail() string {
	return pm.Email
}

// Тема сообщения
func (pm *PlainMail) MessageSubject() string {
	return pm.Subject
}

// Тип контента
func (pm *PlainMail) MessageType() string {
	return "text/plain"
}

// Тело сообщения
func (pm *PlainMail) MessageContent() []byte {
	return []byte(pm.Message)
}
