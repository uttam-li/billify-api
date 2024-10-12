package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTAuthenticator struct {
	accSecret string
	refSecret string
	iss       string
	aud       string
	accExp    time.Duration
	refExp    time.Duration
}

func NewJWTAuthenticator(accSecret, refSecret, iss, aud string, accExp, refExp time.Duration) *JWTAuthenticator {
	return &JWTAuthenticator{
		accSecret: accSecret,
		refSecret: refSecret,
		iss:       iss,
		aud:       aud,
		accExp:    accExp,
		refExp:    refExp,
	}
}

func (a *JWTAuthenticator) GenerateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(a.accExp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": a.iss,
		"aud": a.aud,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.accSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *JWTAuthenticator) ValidateAccessToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(a.accSecret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}

func (a *JWTAuthenticator) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(a.refExp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": a.iss,
		"aud": a.aud,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.refSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *JWTAuthenticator) ValidateRefreshToken(token string) (*jwt.Token, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(a.refSecret), nil
	},
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return parsedToken, nil
}

func (a *JWTAuthenticator) RefreshToken(refreshTokenString string) (string, error) {
	token, err := a.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("unexpected claims type")
	}

	claims["exp"] = time.Now().Add(a.accExp).Unix()

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := newToken.SignedString([]byte(a.accSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}