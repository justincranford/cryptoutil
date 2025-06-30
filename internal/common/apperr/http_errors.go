package apperr

import (
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
)

type Error struct {
	Timestamp             time.Time             // ISO 8601 UTC format timestamp
	ID                    googleUuid.UUID       // Correlation ID (User facing message <=> Internal telemetry log)
	HTTPStatusLineAndCode HTTPStatusLineAndCode // HTTP status code and HTTP reason phrase
	Summary               string                // User-facing summary message
	Err                   error                 // Optional error
}

func (e *Error) Error() string {
	timestamp := e.Timestamp.UTC().Format(time.RFC3339Nano)
	id := e.ID.String()
	code := e.HTTPStatusLineAndCode.Code
	summary := e.Summary
	var err string
	if e.Err == nil {
		err = fmt.Sprintf("%s %s %s %s", timestamp, code, summary, id)
	} else {
		err = fmt.Sprintf("[%s] [%s] [%s] [%s]: %v", timestamp, code, summary, id, e.Err)
	}
	return err
}

func New(httpStatusLineAndCode *HTTPStatusLineAndCode, summary *string, err error) *Error {
	return &Error{
		ID:                    googleUuid.Must(googleUuid.NewV7()),
		HTTPStatusLineAndCode: *httpStatusLineAndCode,
		Summary:               *summary,
		Err:                   err,
		Timestamp:             time.Now().UTC(),
	}
}

func NewHTTP400BadRequest(summary *string, err error) *Error {
	return New(&HTTP400StatusLineAndCodeBadRequest, summary, err)
}

func NewHTTP401Unauthorized(summary *string, err error) *Error {
	return New(&HTTP401StatusLineAndCodeUnauthorized, summary, err)
}

func NewHTTP402PaymentRequired(summary *string, err error) *Error {
	return New(&HTTP402StatusLineAndCodePaymentRequired, summary, err)
}

func NewHTTP403Forbidden(summary *string, err error) *Error {
	return New(&HTTP403StatusLineAndCodeForbidden, summary, err)
}

func NewHTTP404NotFound(summary *string, err error) *Error {
	return New(&HTTP404StatusLineAndCodeNotFound, summary, err)
}

func NewHTTP405MethodNotAllowed(summary *string, err error) *Error {
	return New(&HTTP405StatusLineAndCodeMethodNotAllowed, summary, err)
}

func NewHTTP406NotAcceptable(summary *string, err error) *Error {
	return New(&HTTP406StatusLineAndCodeNotAcceptable, summary, err)
}

func NewHTTP407ProxyAuthRequired(summary *string, err error) *Error {
	return New(&HTTP407StatusLineAndCodeProxyAuthRequired, summary, err)
}

func NewHTTP408RequestTimeout(summary *string, err error) *Error {
	return New(&HTTP408StatusLineAndCodeRequestTimeout, summary, err)
}

func NewHTTP409Conflict(summary *string, err error) *Error {
	return New(&HTTP409StatusLineAndCodeConflict, summary, err)
}

func NewHTTP410Gone(summary *string, err error) *Error {
	return New(&HTTP410StatusLineAndCodeGone, summary, err)
}

func NewHTTP411LengthRequired(summary *string, err error) *Error {
	return New(&HTTP411StatusLineAndCodeLengthRequired, summary, err)
}

func NewHTTP412PreconditionFailed(summary *string, err error) *Error {
	return New(&HTTP412StatusLineAndCodePreconditionFailed, summary, err)
}

func NewHTTP413PayloadTooLarge(summary *string, err error) *Error {
	return New(&HTTP413StatusLineAndCodePayloadTooLarge, summary, err)
}

func NewHTTP414URITooLong(summary *string, err error) *Error {
	return New(&HTTP414StatusLineAndCodeURITooLong, summary, err)
}

func NewHTTP415UnsupportedMediaType(summary *string, err error) *Error {
	return New(&HTTP415StatusLineAndCodeUnsupportedMediaType, summary, err)
}

func NewHTTP416RangeNotSatisfiable(summary *string, err error) *Error {
	return New(&HTTP416StatusLineAndCodeRangeNotSatisfiable, summary, err)
}

func NewHTTP417ExpectationFailed(summary *string, err error) *Error {
	return New(&HTTP417StatusLineAndCodeExpectationFailed, summary, err)
}

func NewHTTP418Teapot(summary *string, err error) *Error {
	return New(&HTTP418StatusLineAndCodeTeapot, summary, err)
}

func NewHTTP421MisdirectedRequest(summary *string, err error) *Error {
	return New(&HTTP421StatusLineAndCodeMisdirectedRequest, summary, err)
}

func NewHTTP422UnprocessableEntity(summary *string, err error) *Error {
	return New(&HTTP422StatusLineAndCodeUnprocessableEntity, summary, err)
}

func NewHTTP423Locked(summary *string, err error) *Error {
	return New(&HTTP423StatusLineAndCodeLocked, summary, err)
}

func NewHTTP424FailedDependency(summary *string, err error) *Error {
	return New(&HTTP424StatusLineAndCodeFailedDependency, summary, err)
}

func NewHTTP425TooEarly(summary *string, err error) *Error {
	return New(&HTTP425StatusLineAndCodeTooEarly, summary, err)
}

func NewHTTP426UpgradeRequired(summary *string, err error) *Error {
	return New(&HTTP426StatusLineAndCodeUpgradeRequired, summary, err)
}

func NewHTTP428PreconditionRequired(summary *string, err error) *Error {
	return New(&HTTP428StatusLineAndCodePreconditionRequired, summary, err)
}

func NewHTTP429TooManyRequests(summary *string, err error) *Error {
	return New(&HTTP429StatusLineAndCodeTooManyRequests, summary, err)
}

func NewHTTP431RequestHeaderFieldsTooLarge(summary *string, err error) *Error {
	return New(&HTTP431StatusLineAndCodeRequestHeaderFieldsTooLarge, summary, err)
}

func NewHTTP451UnavailableForLegalReasons(summary *string, err error) *Error {
	return New(&HTTP451StatusLineAndCodeUnavailableForLegalReasons, summary, err)
}

// 5xx

func NewHTTP500InternalServerError(summary *string, err error) *Error {
	return New(&HTTP500StatusLineAndCodeInternalServerError, summary, err)
}

func NewHTTP501StatusLineAndCodeNotImplemented(summary *string, err error) *Error {
	return New(&HTTP501StatusLineAndCodeNotImplemented, summary, err)
}

func NewHTTP502StatusLineAndCodeBadGateway(summary *string, err error) *Error {
	return New(&HTTP502StatusLineAndCodeBadGateway, summary, err)
}

func NewHTTP503StatusLineAndCodeServiceUnavailable(summary *string, err error) *Error {
	return New(&HTTP503StatusLineAndCodeServiceUnavailable, summary, err)
}

func NewHTTP504StatusLineAndCodeGatewayTimeout(summary *string, err error) *Error {
	return New(&HTTP504StatusLineAndCodeGatewayTimeout, summary, err)
}

func NewHTTP505StatusLineAndCodeHTTPVersionNotSupported(summary *string, err error) *Error {
	return New(&HTTP505StatusLineAndCodeHTTPVersionNotSupported, summary, err)
}

func NewHTTP506StatusLineAndCodeVariantAlsoNegotiates(summary *string, err error) *Error {
	return New(&HTTP506StatusLineAndCodeVariantAlsoNegotiates, summary, err)
}

func NewHTTP507StatusLineAndCodeInsufficientStorage(summary *string, err error) *Error {
	return New(&HTTP507StatusLineAndCodeInsufficientStorage, summary, err)
}

func NewHTTP508StatusLineAndCodeLoopDetected(summary *string, err error) *Error {
	return New(&HTTP508StatusLineAndCodeLoopDetected, summary, err)
}

func NewHTTP510StatusLineAndCodeNotExtended(summary *string, err error) *Error {
	return New(&HTTP510StatusLineAndCodeNotExtended, summary, err)
}

func NewHTTP511StatusLineAndCodeNetworkAuthenticationRequired(summary *string, err error) *Error {
	return New(&HTTP511StatusLineAndCodeNetworkAuthenticationRequired, summary, err)
}
