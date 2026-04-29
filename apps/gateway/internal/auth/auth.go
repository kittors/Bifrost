package auth

import "errors"

var (
	ErrInvalidPasswordHash = errors.New("invalid password hash")
	ErrInvalidToken        = errors.New("invalid token")
	ErrExpiredToken        = errors.New("token expired")
)
