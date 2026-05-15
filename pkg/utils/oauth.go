package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"vault/be/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

var (
	GoogleOauthConfig   *oauth2.Config
	FacebookOauthConfig *oauth2.Config
)

func InitOauth() {
	GoogleOauthConfig = &oauth2.Config{
		RedirectURL:  config.App.GoogleRedirectURI,
		ClientID:     config.App.GoogleClientID,
		ClientSecret: config.App.GoogleClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	FacebookOauthConfig = &oauth2.Config{
		RedirectURL:  config.App.FacebookRedirectURI,
		ClientID:     config.App.FacebookAppID,
		ClientSecret: config.App.FacebookAppSecret,
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func GetGoogleUserInfo(code string) (*GoogleUser, error) {
	token, err := GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	var user GoogleUser
	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed decoding user info: %s", err.Error())
	}

	return &user, nil
}

type FacebookUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

func GetFacebookUserInfo(code string) (*FacebookUser, error) {
	token, err := FacebookOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	var user FacebookUser
	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed decoding user info: %s", err.Error())
	}

	return &user, nil
}
