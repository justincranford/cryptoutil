package apperror

import (
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
)

type Summary string
type Details string
type Log string

type Error struct {
	Timestamp             time.Time             // ISO 8601 UTC format timestamp
	ID                    googleUuid.UUID       // Correlation ID (User facing message <=> Internal telemetry log)
	HTTPStatusLineAndCode HTTPStatusLineAndCode // HTTP status code and HTTP reason phrase
	Summary               Summary               // User-facing summary message
	Details               Details               // User-facing detailed message
	Log                   Log                   // Internal-facing log message
}

func New(httpStatusLineAndCode HTTPStatusLineAndCode, summary Summary, details Details, log Log) *Error {
	return &Error{
		ID:                    googleUuid.Must(googleUuid.NewV7()),
		HTTPStatusLineAndCode: httpStatusLineAndCode,
		Summary:               summary,
		Details:               details,
		Log:                   log,
		Timestamp:             time.Now().UTC(),
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s %s %s %s", e.Timestamp.UTC().Format(time.RFC3339Nano), e.HTTPStatusLineAndCode.Code, e.Summary, e.ID.String())
}

var (
	Err400BadRequest = New(
		HTTP400StatusLineAndCodeBadRequest,
		"Invalid request syntax.",
		"The request could not be understood due to malformed syntax.",
		"Client sent an invalid request.",
	)

	Err404NotFound = New(
		HTTP404StatusLineAndCodeNotFound,
		"The requested resource was not found.",
		"The server could not find the requested resource.",
		"Requested endpoint is missing.",
	)

	Err500InternalServerError = New(
		HTTP500StatusLineAndCodeInternalServerError,
		"An internal server error occurred.",
		"Unexpected error encountered by the server.",
		"Unhandled exception in server code.",
	)
)
