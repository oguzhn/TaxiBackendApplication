package authentication

import (
	"errors"
	"net/http"
)

type HardCodedAuthenticator struct {
	token    string
	client   Client
	username string
	password string
}

func NewHardCodedAuthenticator(token, username, password string) HardCodedAuthenticator {
	return HardCodedAuthenticator{token: token, username: username, password: password, client: admin{}}
}

func (auth HardCodedAuthenticator) Authenticate(r *http.Request) (Client, error) {
	if r.Header.Get("Authorization") != auth.token {
		return nil, errors.New("Unauthorized client")
	}

	return auth.client, nil
}

func (auth HardCodedAuthenticator) Login(username, password string) (string, error) {
	if auth.username != username || auth.password != password {
		return "", errors.New("wrong username or password")
	}
	return auth.token, nil
}

type Loginer interface {
	Login(string, string) (string, error)
}

type Authenticator interface {
	Authenticate(r *http.Request) (Client, error)
}

type Client interface {
	Role() ClientRole
}

type ClientRole string

const (
	RoleAdmin ClientRole = "admin"
)

type admin struct {
}

func (a admin) Role() ClientRole {
	return RoleAdmin
}
