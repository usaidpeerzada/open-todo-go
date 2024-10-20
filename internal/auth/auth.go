package auth

import "github.com/golang-jwt/jwt/v5"

type Authenticator interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
	VerifyPassword(token string, password string) bool
}
