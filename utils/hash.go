package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(hash))

	if err != nil {
		return false, err
	}

	return true, nil
}
