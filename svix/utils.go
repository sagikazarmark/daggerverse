package main

import (
	"crypto/rand"
	"crypto/sha1"
	"dagger/svix/internal/dagger"
	"fmt"
	"math/big"
)

func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/~"

	password := make([]byte, length)
	for i := range password {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[randomIndex.Int64()]
	}

	return string(password), nil
}

func generateRandomSecret(prefix string, length int) (*dagger.Secret, error) {
	secret, err := generateRandomString(20)
	if err != nil {
		return nil, err
	}

	h := sha1.New()

	_, err = h.Write([]byte(secret))
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%s-%x", prefix, h.Sum(nil))

	return dag.SetSecret(name, secret), nil
}
