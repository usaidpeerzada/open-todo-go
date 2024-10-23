package main

import (
	"fmt"
	"net/http"
	"open-todo-go/internal/store"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=20"`
	Email    string `json:"email" validate:"required,max=200"`
	Password string `json:"password" validate:"required,min=6,max=30"`
}

type LoggedInUser struct {
	*store.User
	Token string `json:"token"`
}

func (app *application) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		fmt.Println("readJSON", err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		fmt.Println("validation", err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}
	fmt.Println("user", user)

	if err := user.Password.SetPassword(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		fmt.Println("password", err)
		return
	}
	fmt.Println("after_user", user)

	ctx := r.Context()

	// plainToken := uuid.New().String()

	// hash := sha256.Sum256([]byte(plainToken))
	// hashToken := hex.EncodeToString(hash[:])

	err := app.store.Users.Create(ctx, user)
	fmt.Println("create()", err)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}

		return
	}
	app.jsonResponse(w, http.StatusOK, nil)
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

func (app *application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the request payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Get user by email
	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedErrorResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	var userID int64
	// Verify password using the hash stored in user.Password.hash
	if !app.authenticator.VerifyPassword(payload.Password, string(user.Password.Hash)) {
		app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid email or password"))
		return
	} else {
		userID = user.ID
	}

	// Create JWT claims
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.iss,
	}

	// Generate JWT token
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Send the token in response
	if err := app.jsonResponse(w, http.StatusOK, map[string]string{"token": token, "userID": strconv.FormatInt(userID, 10)}); err != nil {
		app.internalServerError(w, r, err)
	}
}
