package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Authenticator interface {
	GenerateAccessToken(userID uuid.UUID) (string, error)
	ValidateAccessToken(token string) (*jwt.Token, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateRefreshToken(token string) (*jwt.Token, error)
	RefreshToken(token string) (string, error)
}

type OAuthConfigReader interface {
    GetClientID() string
    GetClientSecret() string
    GetAuthURL() string
    GetTokenURL() string
    GetRedirectURL() string
    GetScopes() []string
}