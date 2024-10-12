package main

import (
	"billify-api/internal/store"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GenerateRandomState() (string, error) {
	b := make([]byte, 16) // 16 bytes = 128 bits, good enough for randomness
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (app *application) SetRefreshTokenCookie(w http.ResponseWriter, refreshToken string, domain string) {
	
	cookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(app.config.auth.token.refExp),
	}

	http.SetCookie(w, cookie)
}

func (app *application) providerOAuthHandler(w http.ResponseWriter, r *http.Request) {   
	// Get provider from the query parameter
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		http.Error(w, "missing provider", http.StatusBadRequest)
		return
	}

	// Generate the auth URL with state for CSRF protection
	state, err := GenerateRandomState()
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		MaxAge:   30, // 5 minutes
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(app.config.auth.token.refExp),
	})

	url := app.oauth.GetAuthURL(provider, state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type OAuthTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	IDToken      string    `json:"id_token"`
	Expiry       time.Time `json:"expiry"`
}

func (app *application) callbackOAuthHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		http.Error(w, "missing provider", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	cookie, err := r.Cookie("oauthstate")
	if err != nil {
		http.Error(w, "missing state cookie", http.StatusBadRequest)
		return
	}

	if state != cookie.Value {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	token, err := app.oauth.ExchangeToken(provider, code)
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	// TODO: Save the token to the database
	userInfo, err := app.oauth.GetUserInfo(provider, token.AccessToken)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	user, err := app.store.Users.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		if err == store.ErrNotFound {
			if err := app.store.Users.Create(ctx, userInfo); err != nil {
				app.internalServerError(w, r, err)
				return
			}
		} else {
			app.internalServerError(w, r, err)
			return
		}
	}

	userID := userInfo.ID
	if userID == uuid.Nil {
		userID = user.ID
	}
	// Conver expires_in to expires_at
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	oauthProvider := &store.OAuthProvider{
		UserID:       userID,
		Provider:     provider,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    expiresAt,
	}

	if err := app.store.OAuthProvider.CreateOrUpdate(ctx, oauthProvider); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	refreshToken, err := app.token.GenerateRefreshToken(userID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.SetRefreshTokenCookie(w, refreshToken, app.config.frontendURL)

	redirectURL := app.config.frontendURL + "/buss"

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)

}

type RegisterUserPayload struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=255"`
	LastName  string `json:"last_name" validate:"required,min=2,max=255"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=3,max=72"`
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	err := app.store.Users.Create(ctx, user)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.conflictResponse(w, r, errors.New("user with this email already exists"))
		case store.ErrDuplicateUsername:
			app.conflictResponse(w, r, errors.New("username already exists"))
		default:
			app.internalServerError(w, r, err)
		}
	}

	app.jsonResponse(w, http.StatusCreated, user)
}

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required"`
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := user.Password.ComparePassword(payload.Password); err != nil {
		app.unauthorizedErrorResponse(w, r, errors.New("invalid credentials"))
		return
	}

	accessToken, err := app.token.GenerateAccessToken(user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	refreshToken, err := app.token.GenerateRefreshToken(user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.SetRefreshTokenCookie(w, refreshToken, app.config.frontendURL)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"accessToken": "%s"}`, accessToken)))
}

func (app *application) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		app.unauthorizedErrorResponse(w, r, fmt.Errorf("refresh token is missing"))
		return
	}

	refreshToken := cookie.Value
	fmt.Println(refreshToken)
	newAccessToken, err := app.token.RefreshToken(refreshToken)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"accessToken": "%s"}`, newAccessToken)))
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-time.Hour),
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
}