package smtp

import (
	"bytes"
	"errors"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/pkg/helper"
	"html/template"
)

type SendEmailInput struct {
	To      []string
	Subject string
	Body    string
}

type Sender interface {
	Send(input SendEmailInput) error
}

func (e *SendEmailInput) GenerateBodyFromHTML(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		logger.Errorf("failed to parse file %s:%s", templateFileName, err.Error())

		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	e.Body = buf.String()

	return nil
}

func (e *SendEmailInput) Validate() error {
	if len(e.To) == 0 {
		return errors.New("empty to")
	}

	if e.Subject == "" || e.Body == "" {
		return errors.New("empty subject/body")
	}

	for i := range e.To {
		if !helper.IsEmailValid(e.To[i]) {
			return errors.New("invalid to email")
		}
	}

	return nil
}
