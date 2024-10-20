package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTAuthenticator struct {
	secret string
	aud    string
	iss    string
}

func NewJWTAuthenticator(secret, aud, iss string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret,
		iss,
		aud,
	}
}

func (auth *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(auth.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}

		return []byte(a.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.aud),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}

func (a *JWTAuthenticator) VerifyPassword(plainPassword, hashedPassword string) bool {
	// Compare the hashed password with the plain-text password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	fmt.Print(err)
	if err != nil {
		return false
	} else {
		return true
	}
}
