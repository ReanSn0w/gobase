package utils

import (
	"math/rand"
	"os"
	"time"
)

const (
	env     = "SECURE_PHRASE"
	alfabet = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"
	numbers = "1234567890"
	symbols = "!@#$%^&*<>?"
)

var salt string

func init() {
	s, b := os.LookupEnv(env)
	if !b {
		s = generateNewSalt()
	}

	os.Unsetenv(env)

	salt = s
}

// Метод возвращает соль для текущей сессии
func Salt() string {
	return salt
}

func GenerateRandomString(count int, a, n, s bool) string {
	return stringGenerator(count, a, n, s)
}

func generateNewSalt() string {
	return stringGenerator(64, true, true, true)
}

func stringGenerator(count int, a, n, s bool) string {
	chars := ""

	if a {
		chars += alfabet
	}

	if n {
		chars += numbers
	}

	if s {
		chars += symbols
	}

	rand.Seed(time.Now().Unix())
	randLimit := len(chars)
	newSalt := []byte{}

	for i := 0; i < count; i++ {
		index := rand.Intn(randLimit)
		newSalt = append(newSalt, chars[index])
	}

	return string(newSalt)
}
