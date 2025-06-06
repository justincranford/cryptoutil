package datetime

import (
	"testing"
	"time"
)

func TestISO8601Time2String(t *testing.T) {
	// Happy path
	now := time.Now().UTC()
	expected := now.Format(utcFormat)
	result := ISO8601Time2String(&now)
	if result == nil || *result != expected {
		t.Errorf("expected %s, got %s", expected, *result)
	}

	// Sad path
	var nilTime *time.Time = nil
	result = ISO8601Time2String(nilTime)
	if result != nil {
		t.Errorf("expected nil, got %v", *result)
	}
}

func TestISO8601String2Time(t *testing.T) {
	// Happy path
	now := time.Now().UTC().Format(utcFormat)
	expected, err := time.Parse(utcFormat, now)
	if err != nil {
		t.Fatalf("failed to parse date: %v", err)
	}
	result, err := ISO8601String2Time(&now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil || !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Sad path: Invalid format
	invalidDate := "not-a-date"
	result, err = ISO8601String2Time(&invalidDate)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}

	// Sad path: Nil string
	var nilString *string = nil
	result, err = ISO8601String2Time(nilString)
	if err != nil {
		t.Fatalf("expected no error for nil input, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}
