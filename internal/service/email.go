package service

import (
	"context"
	"github.com/begenov/real-estate/internal/config"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/repository/postgres"
	"github.com/begenov/real-estate/pkg/helper"
	"github.com/begenov/real-estate/pkg/smtp"
	"time"
)

type IEmailService interface {
	SendContactForm(name, phone, message string) error
	SendTourForm(name, phone string, desiredDate time.Time) error
}

type EmailService struct {
	sender        smtp.Sender
	cfg           config.EmailConfig
	receiverEmail string
	blockRepo     postgres.IBlockRepo
}

func NewEmailService(sender smtp.Sender, cfg config.EmailConfig, receiverEmail string, blockRepo postgres.IBlockRepo) *EmailService {
	return &EmailService{
		sender:        sender,
		cfg:           cfg,
		receiverEmail: receiverEmail,
		blockRepo:     blockRepo,
	}
}

// SendContactForm отправляет письмо с формы обратной связи
func (s *EmailService) SendContactForm(name, phone, message string) error {
	input := smtp.SendEmailInput{
		Subject: s.cfg.Subjects.ContractFormEmail,
	}

	emails, err := s.blockRepo.GetEmails(context.Background())
	if err != nil {
		logger.Errorf("s.blockRepo.GetEmails() error: %v", err)
		return err
	}

	for i, email := range emails {
		emails[i] = helper.StripHTML(email)
	}

	input.To = append(input.To, emails...)
	input.To = append(input.To, s.receiverEmail)

	data := struct {
		Name    string
		Phone   string
		Message string
	}{
		Name:    name,
		Phone:   phone,
		Message: message,
	}

	err = input.GenerateBodyFromHTML(s.cfg.Templates.ContractFormEmail, data)
	if err != nil {
		return err
	}

	return s.sender.Send(input)
}

func (s *EmailService) SendTourForm(name, phone string, desiredDate time.Time) error {
	input := smtp.SendEmailInput{
		Subject: s.cfg.Subjects.TourFormEmail,
	}

	emails, err := s.blockRepo.GetEmails(context.Background())
	if err != nil {
		logger.Errorf("s.blockRepo.GetEmails() error: %v", err)
		return err
	}

	for i, email := range emails {
		emails[i] = helper.StripHTML(email)
	}

	input.To = append(input.To, emails...)
	input.To = append(input.To, s.receiverEmail)

	data := struct {
		Name        string
		Phone       string
		DesiredDate string
	}{
		Name:        name,
		Phone:       phone,
		DesiredDate: desiredDate.Format(time.DateOnly),
	}

	err = input.GenerateBodyFromHTML(s.cfg.Templates.TourFormEmail, data)
	if err != nil {
		return err
	}

	return s.sender.Send(input)
}
