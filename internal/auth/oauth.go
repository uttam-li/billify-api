package auth

import (
	"billify-api/internal/store"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"gopkg.in/go-jose/go-jose.v2/json"
)

type OAuthAuthenticator struct {
	config map[string]*oauth2.Config
}

type googleUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Link      string `json:"link"`
	Picture   string `json:"picture"`
}

func NewOAuthAuthenticator(configs map[string]OAuthConfigReader) OAuthAuthenticator {
	oauthConfigs := make(map[string]*oauth2.Config)
	for provider, cfg := range configs {
		oauthConfigs[provider] = &oauth2.Config{
			ClientID:     cfg.GetClientID(),
			ClientSecret: cfg.GetClientSecret(),
			RedirectURL:  cfg.GetRedirectURL(),
			Scopes:       cfg.GetScopes(),
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.GetAuthURL(),
				TokenURL: cfg.GetTokenURL(),
			},
		}
	}
	return OAuthAuthenticator{config: oauthConfigs}
}

func (o *OAuthAuthenticator) GetAuthURL(provider, state string) string {
	conf, ok := o.config[provider]
	if !ok {
		log.Printf("Provider %s not found", provider)
		return ""
	}
	return conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (o *OAuthAuthenticator) GetUserInfo(provider, token string) (*store.User, error) {
	var userInfoURL string

	switch provider {
	case "google":
		userInfoURL = fmt.Sprintf("https://www.googleapis.com/oauth2/v3/userinfo?access_token=%s", token)
	default:
		return nil, errors.New("unsupported provider")
	}

	response, err := http.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status code %d", response.StatusCode)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var user googleUser
	if err := json.Unmarshal(responseBytes, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	userInfo := store.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	return &userInfo, nil
}

func (o *OAuthAuthenticator) ExchangeToken(provider, code string) (*oauth2.Token, error) {
	return o.config[provider].Exchange(context.TODO(), code)
}

func (o *OAuthAuthenticator) RefreshToken(provider, refreshToken string) (*oauth2.Token, error) {
	return o.config[provider].TokenSource(context.TODO(), &oauth2.Token{RefreshToken: refreshToken}).Token()
}
