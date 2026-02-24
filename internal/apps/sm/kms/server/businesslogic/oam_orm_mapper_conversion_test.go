package businesslogic

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
)

func TestToOptionalOrmUUIDs(t *testing.T) {
	mapper := NewOamOrmMapper()

	validUUID1 := googleUuid.New()
	validUUID2 := googleUuid.New()
	validUUIDs := []googleUuid.UUID{validUUID1, validUUID2}
	emptyUUIDs := []googleUuid.UUID{}

	tests := []struct {
		name        string
		input       *[]googleUuid.UUID
		expectError bool
		expectNil   bool
	}{
		{"nil input", nil, false, true},
		{"empty slice", &emptyUUIDs, false, true},
		{"valid UUIDs", &validUUIDs, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := mapper.toOptionalOrmUUIDs(tc.input)

			if tc.expectError {
				testify.Error(t, err)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
					testify.Equal(t, *tc.input, result)
				}
			}
		})
	}
}

func TestToOptionalOrmStrings(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	validStrings := []string{"value1", "value2"}
	emptyStrings := []string{}
	stringsWithEmpty := []string{"valid", ""}

	tests := []struct {
		name          string
		input         *[]string
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{"nil input", nil, false, true, ""},
		{"empty slice", &emptyStrings, false, true, ""},
		{"valid strings", &validStrings, false, false, ""},
		{"strings with empty value", &stringsWithEmpty, true, false, "value must not be empty string"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := mapper.toOptionalOrmStrings(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
					testify.Equal(t, *tc.input, result)
				}
			}
		})
	}
}

func TestToOrmDateRange(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	now := time.Now().UTC()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)
	farPast := now.Add(-48 * time.Hour)

	tests := []struct {
		name          string
		minDate       *time.Time
		maxDate       *time.Time
		expectError   bool
		errorContains string
	}{
		{"both nil", nil, nil, false, ""},
		{"valid past range", &farPast, &past, false, ""},
		{"min in future", &future, nil, true, "min date can't be in the future"},
		{"min after max", &past, &farPast, true, "min date must be before max date"},
		{"min equal max", &past, &past, false, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resultMin, resultMax, err := mapper.toOrmDateRange(tc.minDate, tc.maxDate)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)
				testify.Equal(t, tc.minDate, resultMin)
				testify.Equal(t, tc.maxDate, resultMax)
			}
		})
	}
}

func TestToOrmPageNumber(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	zero := cryptoutilKmsServer.PageNumber(0)
	positive := cryptoutilKmsServer.PageNumber(5)
	negative := cryptoutilKmsServer.PageNumber(-1)

	tests := []struct {
		name        string
		input       *cryptoutilKmsServer.PageNumber
		expected    int
		expectError bool
	}{
		{"nil returns default", nil, 0, false},
		{"zero page number", &zero, 0, false},
		{"positive page number", &positive, 5, false},
		{"negative page number", &negative, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := mapper.toOrmPageNumber(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), "page number must be zero or higher")
			} else {
				testify.NoError(t, err)
				testify.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestToOrmPageSize(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	one := cryptoutilKmsServer.PageSize(1)
	ten := cryptoutilKmsServer.PageSize(10)
	zero := cryptoutilKmsServer.PageSize(0)

	tests := []struct {
		name        string
		input       *cryptoutilKmsServer.PageSize
		expectError bool
		minValue    int
	}{
		{"nil returns default", nil, false, 1},
		{"size of one", &one, false, 1},
		{"size of ten", &ten, false, 10},
		{"zero size", &zero, true, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := mapper.toOrmPageSize(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), "page size must be one or higher")
			} else {
				testify.NoError(t, err)
				testify.GreaterOrEqual(t, result, tc.minValue)
			}
		})
	}
}
