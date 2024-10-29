package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func GetBasicCredential(body string) (BasicCredential, error) {
	credential := BasicCredential{}
	err := json.Unmarshal([]byte(body), &credential)
	if err != nil {
		fmt.Println("Couldn't convert request body due to: ", err.Error())
		return BasicCredential{}, err
	}
	return credential, err
}

type BasicCredential struct {
	Username string `json:"username" dynamodbav:"username"`
	Password string `json:"password"`
}

func BuildSessionCookie(provider ServiceProvider) *http.Cookie {
	return &http.Cookie{
		Name:     "session_cookie",
		Value:    base64.StdEncoding.EncodeToString([]byte(provider.Email)),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(8 * time.Hour),
	}
}
