package utils

import (
	"os"

	"github.com/go-chi/jwtauth"
)

func init() {
	setupSalt()

	tokenizer = &jwtUtility{jwt: jwtauth.New("HS256", []byte(Salt()), nil)}
}

func setupSalt() {
	s, b := os.LookupEnv(env)
	if !b {
		s = generateNewSalt()
	}

	os.Unsetenv(env)

	salt = s
}
