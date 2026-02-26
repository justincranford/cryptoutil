// Copyright (c) 2025 Justin Cranford

// Package email provides email delivery services for the identity service.
package email

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailService interface defines email sending capabilities.
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// SMTPConfig holds SMTP server configuration.
type SMTPConfig struct {
	Host     string // SMTP server hostname (e.g., smtp.gmail.com).
	Port     int    // SMTP server port (e.g., 587 for TLS).
	Username string // SMTP authentication username.
	Password string // SMTP authentication password.
	From     string // Sender email address.
}

// SMTPEmailService implements EmailService using SMTP.
type SMTPEmailService struct {
	config SMTPConfig
}

// NewSMTPEmailService creates a new SMTP email service.
func NewSMTPEmailService(config SMTPConfig) *SMTPEmailService {
	return &SMTPEmailService{
		config: config,
	}
}

// SendEmail sends an email via SMTP.
func (s *SMTPEmailService) SendEmail(_ context.Context, to, subject, body string) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.config.From, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	recipients := []string{to}

	if err := smtp.SendMail(addr, auth, s.config.From, recipients, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// MockEmailService implements EmailService for testing.
type MockEmailService struct {
	SentEmails []SentEmail // Records all sent emails for verification.
}

// SentEmail represents an email sent by MockEmailService.
type SentEmail struct {
	To      string
	Subject string
	Body    string
}

// NewMockEmailService creates a new mock email service.
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: []SentEmail{},
	}
}

// SendEmail records the email instead of actually sending it.
func (m *MockEmailService) SendEmail(_ context.Context, to, subject, body string) error {
	m.SentEmails = append(m.SentEmails, SentEmail{
		To:      to,
		Subject: subject,
		Body:    body,
	})

	return nil
}

// GetLastEmail returns the most recently sent email.
func (m *MockEmailService) GetLastEmail() *SentEmail {
	if len(m.SentEmails) == 0 {
		return nil
	}

	return &m.SentEmails[len(m.SentEmails)-1]
}

// ContainsOTP checks if the email body contains a 6-digit OTP.
func (m *MockEmailService) ContainsOTP(email *SentEmail) (string, bool) {
	if email == nil {
		return "", false
	}

	// Extract 6-digit numeric OTP from body.
	words := strings.Fields(email.Body)
	for _, word := range words {
		if len(word) == cryptoutilSharedMagic.DefaultEmailOTPLength && isNumeric(word) {
			return word, true
		}
	}

	return "", false
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
