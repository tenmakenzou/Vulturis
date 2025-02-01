package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/xristoskrik/vulturis/auth"
	"github.com/xristoskrik/vulturis/internal/database"
)

type ApiConfig struct {
	DB        *database.Queries
	SecretKey string
}

func (cfg *ApiConfig) UserCreateHandler(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
		Name string `json:"name"`
		Surname string `json:"surname"`
		Phone string `json:"phone"`
		Mobile string `json:"mobile"`
		Address string `json:"address"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	fmt.Println(params)
	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Can't create user", err)
		return
	}
	user, err := cfg.DB.CreateUser(context.Background(), database.CreateUserParams{
		HashedPassword: hashed,
		Email:          params.Email,
		Name:           params.Name,
		Surname:        params.Surname,
		Phone:          params.Phone,
		Mobile:         params.Mobile,
		Address:        params.Address,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Can't create user", err)
		return
	}

	RespondWithJSON(w, 201, database.User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Name:      user.Name,
		Surname:   user.Surname,
		Phone:     user.Phone,
		Mobile:    user.Mobile,
		Address:   user.Address,
	})
}

func (cfg *ApiConfig) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	//needs email for parameters
	type parameters struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid credentials", err)
		return
	}
	err = cfg.DB.DeleteUserByEmail(context.Background(), params.Email)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Cant find user", err)
		return
	}

	RespondWithJSON(w, http.StatusNoContent, "Successfully deleted user")

}
func (cfg *ApiConfig) UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	//needs email and password or id and email for parameters
	type parameters struct {
		Email    string    `json:"email"`
		Password string    `json:"password"`
		ID       uuid.UUID `json:"id"`
	}
	action := r.URL.Query().Get("action")
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid credentials", err)
		return
	}
	if action == "password" {
		hashed, err := auth.HashPassword(params.Password)

		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Something went wrong!", err)
			return
		}
		_, err = cfg.DB.UpdateUserPasswordByEmail(context.Background(), database.UpdateUserPasswordByEmailParams{
			HashedPassword: hashed,
			Email:          params.Email,
		})
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Cant find user", err)
			return
		}

		RespondWithJSON(w, http.StatusAccepted, "password updated")
		return
	} else if action == "email" {
		_, err = cfg.DB.UpdateUserEmailById(context.Background(), database.UpdateUserEmailByIdParams{
			ID:    params.ID,
			Email: params.Email,
		})
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Cant find user", err)
			return
		}

		RespondWithJSON(w, http.StatusAccepted, "Email updated")
		return
	}

}
func (cfg *ApiConfig) UserGetHandler(w http.ResponseWriter, r *http.Request) {
	//needs id for parameters
	type parameters struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid credentials", err)
		return
	}
	user, err := cfg.DB.GetUser(context.Background(), params.Email)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Cant find user", err)
		return
	}
	RespondWithJSON(w, http.StatusOK, user)
}


func (cfg *ApiConfig) UserloginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		database.User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	fmt.Println(params)
	user, err := cfg.DB.GetUser(context.Background(), params.Email)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "email  wrong", err)
		return
	}
	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, " password wrong", err)
		return
	}
	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.SecretKey,
		time.Hour,
	)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: sql.NullTime{Time: time.Now().AddDate(0, 0, 60), Valid: true},
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    accessToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})
	RespondWithJSON(w, 200, response{
		User: database.User{
				ID:    user.ID,
				Email: user.Email,
				Phone: user.Phone,
				Mobile: user.Mobile,
				Address: user.Address,
				Name: user.Name,
				Surname: user.Surname,			
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})

}
func (cfg *ApiConfig) UserAuthenticateHandler(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		RespondWithError(w, http.StatusUnauthorized, "Missing Authorization header", nil)
		return
	}
	fmt.Println(authHeader)

	// Extract token (assuming Bearer scheme)
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		RespondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format", nil)
		return
	}
	token := tokenParts[1]

	// Validate token
	userID, err := auth.ValidateJWT(token, cfg.SecretKey)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token", err)
		return
	}

	// Fetch user details
	user, err := cfg.DB.GetUserById(context.Background(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "User not found", err)
		return
	}
	fmt.Println(user)
	RespondWithJSON(w, http.StatusOK, database.User{
		ID:    user.ID,
		Email: user.Email,
		Phone: user.Phone,
		Mobile: user.Mobile,
		Address: user.Address,
		Name: user.Name,
		Surname: user.Surname,
	})

}