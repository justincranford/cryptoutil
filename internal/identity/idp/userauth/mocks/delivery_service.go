// Copyright (c) 2025 Justin Cranford
//
//

package mocks

import (
	"context"
	"fmt"
	"sync"
)

// MockSMSProvider implements DeliveryService for testing SMS delivery.
type MockSMSProvider struct {
	mu            sync.RWMutex
	sentMessages  []SMSMessage
	shouldFail    bool
	failureError  error
	callCount     int
	networkErrors map[int]error // Map call count to specific network errors
}

// SMSMessage represents a sent SMS for verification.
type SMSMessage struct {
	PhoneNumber string
	Message     string
	Timestamp   int64
}

// NewMockSMSProvider creates a new mock SMS provider.
func NewMockSMSProvider() *MockSMSProvider {
	return &MockSMSProvider{
		sentMessages:  make([]SMSMessage, 0),
		networkErrors: make(map[int]error),
	}
}

// SendSMS simulates sending an SMS message.
func (m *MockSMSProvider) SendSMS(ctx context.Context, phoneNumber, message string) error {
	// Validate inputs.
	if phoneNumber == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++

	// Check for network error injection at specific call counts.
	if err, exists := m.networkErrors[m.callCount]; exists {
		return err
	}

	// Check for general failure mode.
	if m.shouldFail {
		if m.failureError != nil {
			return m.failureError
		}

		return fmt.Errorf("mock SMS provider configured to fail")
	}

	// Record sent message.
	timestamp := int64(0)

	if ctx != nil {
		if ts, ok := ctx.Value("timestamp").(int64); ok {
			timestamp = ts
		}
	}

	m.sentMessages = append(m.sentMessages, SMSMessage{
		PhoneNumber: phoneNumber,
		Message:     message,
		Timestamp:   timestamp,
	})

	return nil
}

// SendEmail is not implemented for SMS provider (DeliveryService interface requirement).
func (m *MockSMSProvider) SendEmail(_ context.Context, _, _, _ string) error {
	return fmt.Errorf("SendEmail not supported by SMS provider")
}

// SetShouldFail configures the provider to fail all send attempts.
func (m *MockSMSProvider) SetShouldFail(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shouldFail = shouldFail
	m.failureError = err
}

// InjectNetworkError configures the provider to fail at a specific call count.
func (m *MockSMSProvider) InjectNetworkError(callNumber int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.networkErrors[callNumber] = err
}

// GetSentMessages returns all sent SMS messages for verification.
func (m *MockSMSProvider) GetSentMessages() []SMSMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return copy to avoid race conditions.
	messages := make([]SMSMessage, len(m.sentMessages))
	copy(messages, m.sentMessages)

	return messages
}

// GetCallCount returns the number of SendSMS calls.
func (m *MockSMSProvider) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.callCount
}

// Reset clears all sent messages and resets call count.
func (m *MockSMSProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentMessages = make([]SMSMessage, 0)
	m.shouldFail = false
	m.failureError = nil
	m.callCount = 0
	m.networkErrors = make(map[int]error)
}

// MockEmailProvider implements DeliveryService for testing email delivery.
type MockEmailProvider struct {
	mu            sync.RWMutex
	sentEmails    []EmailMessage
	shouldFail    bool
	failureError  error
	callCount     int
	networkErrors map[int]error
}

// EmailMessage represents a sent email for verification.
type EmailMessage struct {
	To        string
	Subject   string
	Body      string
	Timestamp int64
}

// NewMockEmailProvider creates a new mock email provider.
func NewMockEmailProvider() *MockEmailProvider {
	return &MockEmailProvider{
		sentEmails:    make([]EmailMessage, 0),
		networkErrors: make(map[int]error),
	}
}

// SendEmail simulates sending an email message.
func (m *MockEmailProvider) SendEmail(ctx context.Context, to, subject, body string) error {
	// Validate inputs.
	if to == "" {
		return fmt.Errorf("recipient cannot be empty")
	}

	if subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}

	if body == "" {
		return fmt.Errorf("body cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++

	// Check for network error injection.
	if err, exists := m.networkErrors[m.callCount]; exists {
		return err
	}

	// Check for general failure mode.
	if m.shouldFail {
		if m.failureError != nil {
			return m.failureError
		}

		return fmt.Errorf("mock email provider configured to fail")
	}

	// Record sent email.
	timestamp := int64(0)

	if ctx != nil {
		if ts, ok := ctx.Value("timestamp").(int64); ok {
			timestamp = ts
		}
	}

	m.sentEmails = append(m.sentEmails, EmailMessage{
		To:        to,
		Subject:   subject,
		Body:      body,
		Timestamp: timestamp,
	})

	return nil
}

// SendSMS is not implemented for email provider (DeliveryService interface requirement).
func (m *MockEmailProvider) SendSMS(_ context.Context, _, _ string) error {
	return fmt.Errorf("SendSMS not supported by email provider")
}

// SetShouldFail configures the provider to fail all send attempts.
func (m *MockEmailProvider) SetShouldFail(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shouldFail = shouldFail
	m.failureError = err
}

// InjectNetworkError configures the provider to fail at a specific call count.
func (m *MockEmailProvider) InjectNetworkError(callNumber int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.networkErrors[callNumber] = err
}

// GetSentEmails returns all sent emails for verification.
func (m *MockEmailProvider) GetSentEmails() []EmailMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := make([]EmailMessage, len(m.sentEmails))
	copy(messages, m.sentEmails)

	return messages
}

// GetCallCount returns the number of SendEmail calls.
func (m *MockEmailProvider) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.callCount
}

// Reset clears all sent emails and resets call count.
func (m *MockEmailProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentEmails = make([]EmailMessage, 0)
	m.shouldFail = false
	m.failureError = nil
	m.callCount = 0
	m.networkErrors = make(map[int]error)
}
