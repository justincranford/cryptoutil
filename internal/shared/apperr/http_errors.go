// Copyright (c) 2025 Justin Cranford
//
//

package apperr

import (
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
)

// Error represents an HTTP error with correlation ID for debugging.
type Error struct {
	Timestamp             time.Time             // ISO 8601 UTC format timestamp.
	ID                    googleUuid.UUID       // Correlation ID (User facing message <=> Internal telemetry log).
	HTTPStatusLineAndCode HTTPStatusLineAndCode // HTTP status code and HTTP reason phrase.
	Summary               string                // User-facing summary message.
	Err                   error                 // Optional error.
}

// Error implements the error interface for Error.
func (e *Error) Error() string {
	timestamp := e.Timestamp.UTC().Format(time.RFC3339Nano)
	id := e.ID.String()
	code := e.HTTPStatusLineAndCode.Code
	summary := e.Summary

	if e.Err != nil {
		return fmt.Sprintf("[%s] [%s] [%s] [%s]: %v", timestamp, code, summary, id, e.Err)
	}

	return fmt.Sprintf("%s %s %s %s", timestamp, code, summary, id)
}

// New creates a new Error with the given HTTP status, summary, and optional error.
func New(httpStatusLineAndCode *HTTPStatusLineAndCode, summary *string, err error) *Error {
	return &Error{
		ID:                    googleUuid.Must(googleUuid.NewV7()),
		HTTPStatusLineAndCode: *httpStatusLineAndCode,
		Summary:               *summary,
		Err:                   err,
		Timestamp:             time.Now().UTC(),
	}
}

// NewHTTP400BadRequest creates a 400 Bad Request error.
func NewHTTP400BadRequest(summary *string, err error) *Error {
	return New(&HTTP400StatusLineAndCodeBadRequest, summary, err)
}

// NewHTTP401Unauthorized creates a 401 Unauthorized error.
func NewHTTP401Unauthorized(summary *string, err error) *Error {
	return New(&HTTP401StatusLineAndCodeUnauthorized, summary, err)
}

// NewHTTP402PaymentRequired creates a 402 Payment Required error.
func NewHTTP402PaymentRequired(summary *string, err error) *Error {
	return New(&HTTP402StatusLineAndCodePaymentRequired, summary, err)
}

// NewHTTP403Forbidden creates a 403 Forbidden error.
func NewHTTP403Forbidden(summary *string, err error) *Error {
	return New(&HTTP403StatusLineAndCodeForbidden, summary, err)
}

// NewHTTP404NotFound creates a 404 Not Found error.
func NewHTTP404NotFound(summary *string, err error) *Error {
	return New(&HTTP404StatusLineAndCodeNotFound, summary, err)
}

// NewHTTP405MethodNotAllowed creates a 405 Method Not Allowed error.
func NewHTTP405MethodNotAllowed(summary *string, err error) *Error {
	return New(&HTTP405StatusLineAndCodeMethodNotAllowed, summary, err)
}

// NewHTTP406NotAcceptable creates a 406 Not Acceptable error.
func NewHTTP406NotAcceptable(summary *string, err error) *Error {
	return New(&HTTP406StatusLineAndCodeNotAcceptable, summary, err)
}

// NewHTTP407ProxyAuthRequired creates a 407 Proxy Authentication Required error.
func NewHTTP407ProxyAuthRequired(summary *string, err error) *Error {
	return New(&HTTP407StatusLineAndCodeProxyAuthRequired, summary, err)
}

// NewHTTP408RequestTimeout creates a 408 Request Timeout error.
func NewHTTP408RequestTimeout(summary *string, err error) *Error {
	return New(&HTTP408StatusLineAndCodeRequestTimeout, summary, err)
}

// NewHTTP409Conflict creates a 409 Conflict error.
func NewHTTP409Conflict(summary *string, err error) *Error {
	return New(&HTTP409StatusLineAndCodeConflict, summary, err)
}

// NewHTTP410Gone creates a 410 Gone error.
func NewHTTP410Gone(summary *string, err error) *Error {
	return New(&HTTP410StatusLineAndCodeGone, summary, err)
}

// NewHTTP411LengthRequired creates a 411 Length Required error.
func NewHTTP411LengthRequired(summary *string, err error) *Error {
	return New(&HTTP411StatusLineAndCodeLengthRequired, summary, err)
}

// NewHTTP412PreconditionFailed creates a 412 Precondition Failed error.
func NewHTTP412PreconditionFailed(summary *string, err error) *Error {
	return New(&HTTP412StatusLineAndCodePreconditionFailed, summary, err)
}

// NewHTTP413PayloadTooLarge creates a 413 Payload Too Large error.
func NewHTTP413PayloadTooLarge(summary *string, err error) *Error {
	return New(&HTTP413StatusLineAndCodePayloadTooLarge, summary, err)
}

// NewHTTP414URITooLong creates a 414 URI Too Long error.
func NewHTTP414URITooLong(summary *string, err error) *Error {
	return New(&HTTP414StatusLineAndCodeURITooLong, summary, err)
}

// NewHTTP415UnsupportedMediaType creates a 415 Unsupported Media Type error.
func NewHTTP415UnsupportedMediaType(summary *string, err error) *Error {
	return New(&HTTP415StatusLineAndCodeUnsupportedMediaType, summary, err)
}

// NewHTTP416RangeNotSatisfiable creates a 416 Range Not Satisfiable error.
func NewHTTP416RangeNotSatisfiable(summary *string, err error) *Error {
	return New(&HTTP416StatusLineAndCodeRangeNotSatisfiable, summary, err)
}

// NewHTTP417ExpectationFailed creates a 417 Expectation Failed error.
func NewHTTP417ExpectationFailed(summary *string, err error) *Error {
	return New(&HTTP417StatusLineAndCodeExpectationFailed, summary, err)
}

// NewHTTP418Teapot creates a 418 I'm a teapot error.
func NewHTTP418Teapot(summary *string, err error) *Error {
	return New(&HTTP418StatusLineAndCodeTeapot, summary, err)
}

// NewHTTP421MisdirectedRequest creates a 421 Misdirected Request error.
func NewHTTP421MisdirectedRequest(summary *string, err error) *Error {
	return New(&HTTP421StatusLineAndCodeMisdirectedRequest, summary, err)
}

// NewHTTP422UnprocessableEntity creates a 422 Unprocessable Entity error.
func NewHTTP422UnprocessableEntity(summary *string, err error) *Error {
	return New(&HTTP422StatusLineAndCodeUnprocessableEntity, summary, err)
}

// NewHTTP423Locked creates a 423 Locked error.
func NewHTTP423Locked(summary *string, err error) *Error {
	return New(&HTTP423StatusLineAndCodeLocked, summary, err)
}

// NewHTTP424FailedDependency creates a 424 Failed Dependency error.
func NewHTTP424FailedDependency(summary *string, err error) *Error {
	return New(&HTTP424StatusLineAndCodeFailedDependency, summary, err)
}

// NewHTTP425TooEarly creates a 425 Too Early error.
func NewHTTP425TooEarly(summary *string, err error) *Error {
	return New(&HTTP425StatusLineAndCodeTooEarly, summary, err)
}

// NewHTTP426UpgradeRequired creates a 426 Upgrade Required error.
func NewHTTP426UpgradeRequired(summary *string, err error) *Error {
	return New(&HTTP426StatusLineAndCodeUpgradeRequired, summary, err)
}

// NewHTTP428PreconditionRequired creates a 428 Precondition Required error.
func NewHTTP428PreconditionRequired(summary *string, err error) *Error {
	return New(&HTTP428StatusLineAndCodePreconditionRequired, summary, err)
}

// NewHTTP429TooManyRequests creates a 429 Too Many Requests error.
func NewHTTP429TooManyRequests(summary *string, err error) *Error {
	return New(&HTTP429StatusLineAndCodeTooManyRequests, summary, err)
}

// NewHTTP431RequestHeaderFieldsTooLarge creates a 431 Request Header Fields Too Large error.
func NewHTTP431RequestHeaderFieldsTooLarge(summary *string, err error) *Error {
	return New(&HTTP431StatusLineAndCodeRequestHeaderFieldsTooLarge, summary, err)
}

// NewHTTP451UnavailableForLegalReasons creates a 451 Unavailable For Legal Reasons error.
func NewHTTP451UnavailableForLegalReasons(summary *string, err error) *Error {
	return New(&HTTP451StatusLineAndCodeUnavailableForLegalReasons, summary, err)
}

// NewHTTP500InternalServerError creates a 500 Internal Server Error.
func NewHTTP500InternalServerError(summary *string, err error) *Error {
	return New(&HTTP500StatusLineAndCodeInternalServerError, summary, err)
}

// NewHTTP501StatusLineAndCodeNotImplemented creates a 501 Not Implemented error.
func NewHTTP501StatusLineAndCodeNotImplemented(summary *string, err error) *Error {
	return New(&HTTP501StatusLineAndCodeNotImplemented, summary, err)
}

// NewHTTP502StatusLineAndCodeBadGateway creates a 502 Bad Gateway error.
func NewHTTP502StatusLineAndCodeBadGateway(summary *string, err error) *Error {
	return New(&HTTP502StatusLineAndCodeBadGateway, summary, err)
}

// NewHTTP503StatusLineAndCodeServiceUnavailable creates a 503 Service Unavailable error.
func NewHTTP503StatusLineAndCodeServiceUnavailable(summary *string, err error) *Error {
	return New(&HTTP503StatusLineAndCodeServiceUnavailable, summary, err)
}

// NewHTTP504StatusLineAndCodeGatewayTimeout creates a 504 Gateway Timeout error.
func NewHTTP504StatusLineAndCodeGatewayTimeout(summary *string, err error) *Error {
	return New(&HTTP504StatusLineAndCodeGatewayTimeout, summary, err)
}

// NewHTTP505StatusLineAndCodeHTTPVersionNotSupported creates a 505 HTTP Version Not Supported error.
func NewHTTP505StatusLineAndCodeHTTPVersionNotSupported(summary *string, err error) *Error {
	return New(&HTTP505StatusLineAndCodeHTTPVersionNotSupported, summary, err)
}

// NewHTTP506StatusLineAndCodeVariantAlsoNegotiates creates a 506 Variant Also Negotiates error.
func NewHTTP506StatusLineAndCodeVariantAlsoNegotiates(summary *string, err error) *Error {
	return New(&HTTP506StatusLineAndCodeVariantAlsoNegotiates, summary, err)
}

// NewHTTP507StatusLineAndCodeInsufficientStorage creates a 507 Insufficient Storage error.
func NewHTTP507StatusLineAndCodeInsufficientStorage(summary *string, err error) *Error {
	return New(&HTTP507StatusLineAndCodeInsufficientStorage, summary, err)
}

// NewHTTP508StatusLineAndCodeLoopDetected creates a 508 Loop Detected error.
func NewHTTP508StatusLineAndCodeLoopDetected(summary *string, err error) *Error {
	return New(&HTTP508StatusLineAndCodeLoopDetected, summary, err)
}

// NewHTTP510StatusLineAndCodeNotExtended creates a 510 Not Extended error.
func NewHTTP510StatusLineAndCodeNotExtended(summary *string, err error) *Error {
	return New(&HTTP510StatusLineAndCodeNotExtended, summary, err)
}

// NewHTTP511StatusLineAndCodeNetworkAuthenticationRequired creates a 511 Network Authentication Required error.
func NewHTTP511StatusLineAndCodeNetworkAuthenticationRequired(summary *string, err error) *Error {
	return New(&HTTP511StatusLineAndCodeNetworkAuthenticationRequired, summary, err)
}
