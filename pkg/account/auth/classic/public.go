package classic

import (
	"errors"
	"time"

	"github.com/ReanSn0w/gobase/pkg/account"
	"github.com/ReanSn0w/gobase/pkg/account/secure"
	"github.com/ReanSn0w/gobase/pkg/utils"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const (
	emailTokenKey             = "user_email"
	registrationTokenType     = "new_user_registration"
	passwordRecoveryTokenType = "new_password_recovery"
)

var (
	ErrUnvalidToken      = errors.New("токен пользователя истек или непригоден для данного действия")
	ErrAuthentification  = errors.New("email пользователя или пароль не верны")
	ErrEmailNotRegistred = errors.New("данный email не используется ни одним профилем в системе")
)

// Запрос на регистрацию пользователя
//
// Данный запрос создаст токен для подтверждения почты,
// используя данный токен в постледствии пользователь сможет зарегистрироваться на сайте
// Важно! Токен подтверждения Email будет валиден для регистрации 24 часа с момента генерации
func NewRegistrationRequest(email string) (string, error) {
	err := emailAvaliable(email)
	if err != nil {
		return "", err
	}

	return utils.JWT().GenerateToken(jwt.MapClaims{
		"user_email": email,
		"type":       registrationTokenType,
		"iat":        time.Now().Add(time.Hour * 24).Unix(),
	})
}

// Регистрация пользователя
//
// Регистрация пользователя использует токен, который можно получить из функции NewRegistrationRequest()
// функция проверяет, что пользователь с данным email еще не зарегистрирован на сайте
// далее создает новйы профиль для пользователя
// на выходе возвращает токен для авторизации пользователя и интерфейс ошибки
func RegisterUser(token string, password string, session secure.Session) (string, error) {
	claims, err := utils.JWT().Parse(token)
	if err != nil || claims.Valid() != nil || claims["type"] == registrationTokenType {
		return "", ErrUnvalidToken
	}

	email := claims["user_email"].(string)
	err = emailAvaliable(email)
	if err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	userID, err := account.CreateNewAccount(utils.GenerateRandomString(12, true, false, false), "user", session)
	if err != nil {
		return "", err
	}

	err = saveUserCredentials(userID, email, hash)
	if err != nil {
		return "", err
	}

	return utils.JWT().GenerateToken(jwt.MapClaims{
		"user_id":    userID.Hex(),
		"user_group": "user",
		"session":    session.Key,
		"iat":        time.Now().Add(time.Hour),
	})
}

// Авторизация пользователя
//
// Фукция пытается загрузить данные о пользователе из БД и в случае любой возпращает ErrAuthentification
func LoginUser(email string, password string, session secure.Session) (string, error) {
	auth, err := loadUserCredentialsByEmail(email)
	if err != nil || auth == nil {
		return "", ErrAuthentification
	}

	if !auth.Validate(password) {
		return "", ErrAuthentification
	}

	return utils.JWT().GenerateToken(jwt.MapClaims{
		"user_id":    auth.ID.Hex(),
		"user_group": auth.Group,
		"session":    session.Key,
		"iat":        time.Now().Add(time.Hour),
	})
}

// Запрос на восстановление пароля
//
// Токен выдаваемый данной функцией служит для восстановления пароля,
// его следует передать пользователю по email
func NewPasswordReciveryRequest(email string) (string, error) {
	err := emailAvaliable(email)
	if err == nil {
		return "", ErrEmailNotRegistred
	}

	return utils.JWT().GenerateToken(jwt.MapClaims{
		"user_email": email,
		"type":       passwordRecoveryTokenType,
		"iat":        time.Now().Add(time.Hour * 24).Unix(),
	})
}

// Восстановление пароля
//
// Данная функция перезаписывает пароль для авторизации пользователя, в случае успешной валидации токена
func RecoverUserPassword(token string, newPassword string) error {
	claims, err := utils.JWT().Parse(token)
	if err != nil || claims.Valid() != nil || claims["type"] != passwordRecoveryTokenType {
		return ErrUnvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	email := claims["user_email"].(string)

	auth, err := loadUserCredentialsByEmail(email)
	if err != nil {
		return err
	}

	return saveUserCredentials(auth.ID, email, hash)
}

// Функция для изменения пароля и email пользователя
//
// Следует использовать только для зарегистрированных пользователей
// Метод прадставлен для изменения данных входа у пользователей, которые уже залогинены в системе
func ChangeCredentials(userID primitive.ObjectID, email string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return saveUserCredentials(userID, email, hash)
}
