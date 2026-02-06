package smtp

import (
	"errors"
	"net/smtp"
)

type Credentials struct {
	login, password string
}

func Authenticate(login string, password string) *Credentials {
	return &Credentials{login: login, password: password}
}

func (c *Credentials) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (c *Credentials) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(c.login), nil
		case "Password:":
			return []byte(c.password), nil
		default:
			return nil, errors.New("unknown fromServer")
		}
	}
	return nil, nil
}
