package domain

import "errors"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid or expired token")
var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")

type User struct {
	ID          string
	Email       string
	DisplayName string
	Role        string
	Status      string
}

type LoginUser struct {
	User
	PasswordHash string
}
