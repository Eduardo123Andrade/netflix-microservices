package services

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type HashService struct {
	Cost int
}

func NewHasher(cost int) *HashService {
	return &HashService{Cost: cost}
}

var (
	ErrMismatchedHashAndPassword = errors.New("invalid hash and password")
)

func (hs *HashService) Hash(password string) (string, error) {
	pb := ([]byte)(password)

	b, err := bcrypt.GenerateFromPassword(pb, hs.Cost)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (hs *HashService) Compare(hashedValue, value string) error {
	hvb := ([]byte)(hashedValue)
	vb := ([]byte)(value)

	err := bcrypt.CompareHashAndPassword(hvb, vb)

	if err != nil {

		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrMismatchedHashAndPassword
		}

		return err
	}

	return nil

}
