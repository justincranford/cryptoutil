// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"sync"
)

// MockDeliveryService implements DeliveryService for testing.
type MockDeliveryService struct {
	mu         sync.RWMutex
	sentSMS    []SMSMessage
	sentEmails []EmailMessage
	shouldFail bool
}

// SMSMessage represents a sent SMS message.
type SMSMessage struct {
	PhoneNumber string
	Message     string
}

// EmailMessage represents a sent email message.
type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

// NewMockDeliveryService creates a new mock delivery service.
func NewMockDeliveryService() *MockDeliveryService {
	return &MockDeliveryService{
		sentSMS:    make([]SMSMessage, 0),
		sentEmails: make([]EmailMessage, 0),
	}
}

// SendSMS sends an SMS message (mock).
func (m *MockDeliveryService) SendSMS(_ context.Context, phoneNumber, message string) error {
	if m.shouldFail {
		return fmt.Errorf("mock SMS delivery failure")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentSMS = append(m.sentSMS, SMSMessage{
		PhoneNumber: phoneNumber,
		Message:     message,
	})

	return nil
}

// SendEmail sends an email message (mock).
func (m *MockDeliveryService) SendEmail(_ context.Context, to, subject, body string) error {
	if m.shouldFail {
		return fmt.Errorf("mock email delivery failure")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentEmails = append(m.sentEmails, EmailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
	})

	return nil
}

// GetSentSMS returns all sent SMS messages.
func (m *MockDeliveryService) GetSentSMS() []SMSMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]SMSMessage, len(m.sentSMS))
	copy(result, m.sentSMS)

	return result
}

// GetSentEmails returns all sent emails.
func (m *MockDeliveryService) GetSentEmails() []EmailMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]EmailMessage, len(m.sentEmails))
	copy(result, m.sentEmails)

	return result
}

// SetShouldFail sets whether the delivery should fail.
func (m *MockDeliveryService) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shouldFail = shouldFail
}

// Reset clears all sent messages.
func (m *MockDeliveryService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentSMS = make([]SMSMessage, 0)
	m.sentEmails = make([]EmailMessage, 0)
	m.shouldFail = false
}
