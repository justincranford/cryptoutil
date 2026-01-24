// Copyright (c) 2025 Justin Cranford
//
//

package apperr

import http "net/http"

// HTTPStatusCode represents an HTTP status code.
type HTTPStatusCode int

// HTTPReasonPhrase represents an HTTP reason phrase.
type HTTPReasonPhrase string

// ProprietaryAppCode represents a proprietary application error code.
type ProprietaryAppCode string

// HTTPStatusLine represents an HTTP status line with status code and reason phrase.
type HTTPStatusLine struct {
	StatusCode   HTTPStatusCode   // HTTP status code.
	ReasonPhrase HTTPReasonPhrase // HTTP reason phrase.
}

// HTTPStatusLineAndCode represents an HTTP status line with a proprietary application code.
type HTTPStatusLineAndCode struct {
	StatusLine HTTPStatusLine     // HTTP status line (HTTP status code + HTTP reason phrase).
	Code       ProprietaryAppCode // Proprietary Application Code.
}

// NewHTTPStatusLineAndCode creates a new HTTPStatusLineAndCode from a status code and app code.
func NewHTTPStatusLineAndCode(httpStatusCode HTTPStatusCode, proprietartAppCode *ProprietaryAppCode) HTTPStatusLineAndCode {
	return HTTPStatusLineAndCode{
		StatusLine: HTTPStatusLine{
			StatusCode:   httpStatusCode,                                         // 100-599.
			ReasonPhrase: HTTPReasonPhrase(http.StatusText(int(httpStatusCode))), // standard HTTP reason phrase from standard HTTP status codes.
		},
		Code: *proprietartAppCode, // Proprietary application code.
	}
}

// NewHTTPStatusLine creates a new HTTPStatusLine from a status code and reason phrase.
func NewHTTPStatusLine(statusCode HTTPStatusCode, reasonPhrase HTTPReasonPhrase) HTTPStatusLine {
	return HTTPStatusLine{StatusCode: statusCode, ReasonPhrase: reasonPhrase}
}

// NewCode creates a new ProprietaryAppCode from a message string.
func NewCode(message string) ProprietaryAppCode {
	return ProprietaryAppCode(message)
}

// Generate all HTTP status codes (100-599) using net/http constants.
var (
	// HTTP100StatusLineAndCodeContinue represents 100 Continue.
	HTTP100StatusLineAndCodeContinue = NewHTTPStatusLineAndCode(http.StatusContinue, &InfoCodeContinue)
	// HTTP101StatusLineAndCodeSwitchingProtocols represents 101 Switching Protocols.
	HTTP101StatusLineAndCodeSwitchingProtocols = NewHTTPStatusLineAndCode(http.StatusSwitchingProtocols, &InfoCodeSwitchingProtocols)
	// HTTP102StatusLineAndCodeProcessing represents 102 Processing.
	HTTP102StatusLineAndCodeProcessing = NewHTTPStatusLineAndCode(http.StatusProcessing, &InfoCodeProcessing)
	// HTTP103StatusLineAndCodeEarlyHints represents 103 Early Hints.
	HTTP103StatusLineAndCodeEarlyHints = NewHTTPStatusLineAndCode(http.StatusEarlyHints, &InfoCodeEarlyHints)

	// HTTP200StatusLineAndCodeOK represents 200 OK.
	HTTP200StatusLineAndCodeOK = NewHTTPStatusLineAndCode(http.StatusOK, &SuccessCodeOK)
	// HTTP201StatusLineAndCodeCreated represents 201 Created.
	HTTP201StatusLineAndCodeCreated = NewHTTPStatusLineAndCode(http.StatusCreated, &SuccessCodeCreated)
	// HTTP202StatusLineAndCodeAccepted represents 202 Accepted.
	HTTP202StatusLineAndCodeAccepted = NewHTTPStatusLineAndCode(http.StatusAccepted, &SuccessCodeAccepted)
	// HTTP203StatusLineAndCodeNonAuthoritativeInfo represents 203 Non-Authoritative Information.
	HTTP203StatusLineAndCodeNonAuthoritativeInfo = NewHTTPStatusLineAndCode(http.StatusNonAuthoritativeInfo, &SuccessCodeNonAuthoritativeInfo)
	// HTTP204StatusLineAndCodeNoContent represents 204 No Content.
	HTTP204StatusLineAndCodeNoContent = NewHTTPStatusLineAndCode(http.StatusNoContent, &SuccessCodeNoContent)
	// HTTP205StatusLineAndCodeResetContent represents 205 Reset Content.
	HTTP205StatusLineAndCodeResetContent = NewHTTPStatusLineAndCode(http.StatusResetContent, &SuccessCodeResetContent)
	// HTTP206StatusLineAndCodePartialContent represents 206 Partial Content.
	HTTP206StatusLineAndCodePartialContent = NewHTTPStatusLineAndCode(http.StatusPartialContent, &SuccessCodePartialContent)
	// HTTP207StatusLineAndCodeMultiStatus represents 207 Multi-Status.
	HTTP207StatusLineAndCodeMultiStatus = NewHTTPStatusLineAndCode(http.StatusMultiStatus, &SuccessCodeMultiStatus)
	// HTTP208StatusLineAndCodeAlreadyReported represents 208 Already Reported.
	HTTP208StatusLineAndCodeAlreadyReported = NewHTTPStatusLineAndCode(http.StatusAlreadyReported, &SuccessCodeAlreadyReported)
	// HTTP226StatusLineAndCodeIMUsed represents 226 IM Used.
	HTTP226StatusLineAndCodeIMUsed = NewHTTPStatusLineAndCode(http.StatusIMUsed, &SuccessCodeIMUsed)

	// HTTP300StatusLineAndCodeMultipleChoices represents 300 Multiple Choices.
	HTTP300StatusLineAndCodeMultipleChoices = NewHTTPStatusLineAndCode(http.StatusMultipleChoices, &RedirectionCodeMultipleChoices)
	// HTTP301StatusLineAndCodeMovedPermanently represents 301 Moved Permanently.
	HTTP301StatusLineAndCodeMovedPermanently = NewHTTPStatusLineAndCode(http.StatusMovedPermanently, &RedirectionCodeMovedPermanently)
	// HTTP302StatusLineAndCodeFound represents 302 Found.
	HTTP302StatusLineAndCodeFound = NewHTTPStatusLineAndCode(http.StatusFound, &RedirectionCodeFound)
	// HTTP303StatusLineAndCodeSeeOther represents 303 See Other.
	HTTP303StatusLineAndCodeSeeOther = NewHTTPStatusLineAndCode(http.StatusSeeOther, &RedirectionCodeSeeOther)
	// HTTP304StatusLineAndCodeNotModified represents 304 Not Modified.
	HTTP304StatusLineAndCodeNotModified = NewHTTPStatusLineAndCode(http.StatusNotModified, &RedirectionCodeNotModified)
	// HTTP305StatusLineAndCodeUseProxy represents 305 Use Proxy.
	HTTP305StatusLineAndCodeUseProxy = NewHTTPStatusLineAndCode(http.StatusUseProxy, &RedirectionCodeUseProxy)
	// HTTP307StatusLineAndCodeTemporaryRedirect represents 307 Temporary Redirect.
	HTTP307StatusLineAndCodeTemporaryRedirect = NewHTTPStatusLineAndCode(http.StatusTemporaryRedirect, &RedirectionCodeTemporaryRedirect)
	// HTTP308StatusLineAndCodePermanentRedirect represents 308 Permanent Redirect.
	HTTP308StatusLineAndCodePermanentRedirect = NewHTTPStatusLineAndCode(http.StatusPermanentRedirect, &RedirectionCodePermanentRedirect)

	// HTTP400StatusLineAndCodeBadRequest represents 400 Bad Request.
	HTTP400StatusLineAndCodeBadRequest = NewHTTPStatusLineAndCode(http.StatusBadRequest, &ClientErrorCodeBadRequest)
	// HTTP401StatusLineAndCodeUnauthorized represents 401 Unauthorized.
	HTTP401StatusLineAndCodeUnauthorized = NewHTTPStatusLineAndCode(http.StatusUnauthorized, &ClientErrorCodeUnauthorized)
	// HTTP402StatusLineAndCodePaymentRequired represents 402 Payment Required.
	HTTP402StatusLineAndCodePaymentRequired = NewHTTPStatusLineAndCode(http.StatusPaymentRequired, &ClientErrorCodePaymentRequired)
	// HTTP403StatusLineAndCodeForbidden represents 403 Forbidden.
	HTTP403StatusLineAndCodeForbidden = NewHTTPStatusLineAndCode(http.StatusForbidden, &ClientErrorCodeForbidden)
	// HTTP404StatusLineAndCodeNotFound represents 404 Not Found.
	HTTP404StatusLineAndCodeNotFound = NewHTTPStatusLineAndCode(http.StatusNotFound, &ClientErrorCodeNotFound)
	// HTTP405StatusLineAndCodeMethodNotAllowed represents 405 Method Not Allowed.
	HTTP405StatusLineAndCodeMethodNotAllowed = NewHTTPStatusLineAndCode(http.StatusMethodNotAllowed, &ClientErrorCodeMethodNotAllowed)
	// HTTP406StatusLineAndCodeNotAcceptable represents 406 Not Acceptable.
	HTTP406StatusLineAndCodeNotAcceptable = NewHTTPStatusLineAndCode(http.StatusNotAcceptable, &ClientErrorCodeNotAcceptable)
	// HTTP407StatusLineAndCodeProxyAuthRequired represents 407 Proxy Authentication Required.
	HTTP407StatusLineAndCodeProxyAuthRequired = NewHTTPStatusLineAndCode(http.StatusProxyAuthRequired, &ClientErrorCodeProxyAuthRequired)
	// HTTP408StatusLineAndCodeRequestTimeout represents 408 Request Timeout.
	HTTP408StatusLineAndCodeRequestTimeout = NewHTTPStatusLineAndCode(http.StatusRequestTimeout, &ClientErrorCodeRequestTimeout)
	// HTTP409StatusLineAndCodeConflict represents 409 Conflict.
	HTTP409StatusLineAndCodeConflict = NewHTTPStatusLineAndCode(http.StatusConflict, &ClientErrorCodeConflict)
	// HTTP410StatusLineAndCodeGone represents 410 Gone.
	HTTP410StatusLineAndCodeGone = NewHTTPStatusLineAndCode(http.StatusGone, &ClientErrorCodeGone)
	// HTTP411StatusLineAndCodeLengthRequired represents 411 Length Required.
	HTTP411StatusLineAndCodeLengthRequired = NewHTTPStatusLineAndCode(http.StatusLengthRequired, &ClientErrorCodeLengthRequired)
	// HTTP412StatusLineAndCodePreconditionFailed represents 412 Precondition Failed.
	HTTP412StatusLineAndCodePreconditionFailed = NewHTTPStatusLineAndCode(http.StatusPreconditionFailed, &ClientErrorCodePreconditionFailed)
	// HTTP413StatusLineAndCodePayloadTooLarge represents 413 Payload Too Large.
	HTTP413StatusLineAndCodePayloadTooLarge = NewHTTPStatusLineAndCode(http.StatusRequestEntityTooLarge, &ClientErrorCodePayloadTooLarge)
	// HTTP414StatusLineAndCodeURITooLong represents 414 URI Too Long.
	HTTP414StatusLineAndCodeURITooLong = NewHTTPStatusLineAndCode(http.StatusRequestURITooLong, &ClientErrorCodeURITooLong)
	// HTTP415StatusLineAndCodeUnsupportedMediaType represents 415 Unsupported Media Type.
	HTTP415StatusLineAndCodeUnsupportedMediaType = NewHTTPStatusLineAndCode(http.StatusUnsupportedMediaType, &ClientErrorCodeUnsupportedMediaType)
	// HTTP416StatusLineAndCodeRangeNotSatisfiable represents 416 Range Not Satisfiable.
	HTTP416StatusLineAndCodeRangeNotSatisfiable = NewHTTPStatusLineAndCode(http.StatusRequestedRangeNotSatisfiable, &ClientErrorCodeRangeNotSatisfiable)
	// HTTP417StatusLineAndCodeExpectationFailed represents 417 Expectation Failed.
	HTTP417StatusLineAndCodeExpectationFailed = NewHTTPStatusLineAndCode(http.StatusExpectationFailed, &ClientErrorCodeExpectationFailed)
	// HTTP418StatusLineAndCodeTeapot represents 418 I'm a teapot.
	HTTP418StatusLineAndCodeTeapot = NewHTTPStatusLineAndCode(http.StatusTeapot, &ClientErrorCodeTeapot)
	// HTTP421StatusLineAndCodeMisdirectedRequest represents 421 Misdirected Request.
	HTTP421StatusLineAndCodeMisdirectedRequest = NewHTTPStatusLineAndCode(http.StatusMisdirectedRequest, &ClientErrorCodeMisdirectedRequest)
	// HTTP422StatusLineAndCodeUnprocessableEntity represents 422 Unprocessable Entity.
	HTTP422StatusLineAndCodeUnprocessableEntity = NewHTTPStatusLineAndCode(http.StatusUnprocessableEntity, &ClientErrorCodeUnprocessableEntity)
	// HTTP423StatusLineAndCodeLocked represents 423 Locked.
	HTTP423StatusLineAndCodeLocked = NewHTTPStatusLineAndCode(http.StatusLocked, &ClientErrorCodeLocked)
	// HTTP424StatusLineAndCodeFailedDependency represents 424 Failed Dependency.
	HTTP424StatusLineAndCodeFailedDependency = NewHTTPStatusLineAndCode(http.StatusFailedDependency, &ClientErrorCodeFailedDependency)
	// HTTP425StatusLineAndCodeTooEarly represents 425 Too Early.
	HTTP425StatusLineAndCodeTooEarly = NewHTTPStatusLineAndCode(http.StatusTooEarly, &ClientErrorCodeTooEarly)
	// HTTP426StatusLineAndCodeUpgradeRequired represents 426 Upgrade Required.
	HTTP426StatusLineAndCodeUpgradeRequired = NewHTTPStatusLineAndCode(http.StatusUpgradeRequired, &ClientErrorCodeUpgradeRequired)
	// HTTP428StatusLineAndCodePreconditionRequired represents 428 Precondition Required.
	HTTP428StatusLineAndCodePreconditionRequired = NewHTTPStatusLineAndCode(http.StatusPreconditionRequired, &ClientErrorCodePreconditionRequired)
	// HTTP429StatusLineAndCodeTooManyRequests represents 429 Too Many Requests.
	HTTP429StatusLineAndCodeTooManyRequests = NewHTTPStatusLineAndCode(http.StatusTooManyRequests, &ClientErrorCodeTooManyRequests)
	// HTTP431StatusLineAndCodeRequestHeaderFieldsTooLarge represents 431 Request Header Fields Too Large.
	HTTP431StatusLineAndCodeRequestHeaderFieldsTooLarge = NewHTTPStatusLineAndCode(http.StatusRequestHeaderFieldsTooLarge, &ClientErrorCodeRequestHeaderFieldsTooLarge)
	// HTTP451StatusLineAndCodeUnavailableForLegalReasons represents 451 Unavailable For Legal Reasons.
	HTTP451StatusLineAndCodeUnavailableForLegalReasons = NewHTTPStatusLineAndCode(http.StatusUnavailableForLegalReasons, &ClientErrorCodeUnavailableForLegalReasons)

	// HTTP500StatusLineAndCodeInternalServerError represents 500 Internal Server Error.
	HTTP500StatusLineAndCodeInternalServerError = NewHTTPStatusLineAndCode(http.StatusInternalServerError, &ServerErrorCodeInternalServerServerError)
	// HTTP501StatusLineAndCodeNotImplemented represents 501 Not Implemented.
	HTTP501StatusLineAndCodeNotImplemented = NewHTTPStatusLineAndCode(http.StatusNotImplemented, &ServerErrorCodeNotImplemented)
	// HTTP502StatusLineAndCodeBadGateway represents 502 Bad Gateway.
	HTTP502StatusLineAndCodeBadGateway = NewHTTPStatusLineAndCode(http.StatusBadGateway, &ServerErrorCodeBadGateway)
	// HTTP503StatusLineAndCodeServiceUnavailable represents 503 Service Unavailable.
	HTTP503StatusLineAndCodeServiceUnavailable = NewHTTPStatusLineAndCode(http.StatusServiceUnavailable, &ServerErrorCodeServiceUnavailable)
	// HTTP504StatusLineAndCodeGatewayTimeout represents 504 Gateway Timeout.
	HTTP504StatusLineAndCodeGatewayTimeout = NewHTTPStatusLineAndCode(http.StatusGatewayTimeout, &ServerErrorCodeGatewayTimeout)
	// HTTP505StatusLineAndCodeHTTPVersionNotSupported represents 505 HTTP Version Not Supported.
	HTTP505StatusLineAndCodeHTTPVersionNotSupported = NewHTTPStatusLineAndCode(http.StatusHTTPVersionNotSupported, &ServerErrorCodeHTTPVersionNotSupported)
	// HTTP506StatusLineAndCodeVariantAlsoNegotiates represents 506 Variant Also Negotiates.
	HTTP506StatusLineAndCodeVariantAlsoNegotiates = NewHTTPStatusLineAndCode(http.StatusVariantAlsoNegotiates, &ServerErrorCodeVariantAlsoNegotiates)
	// HTTP507StatusLineAndCodeInsufficientStorage represents 507 Insufficient Storage.
	HTTP507StatusLineAndCodeInsufficientStorage = NewHTTPStatusLineAndCode(http.StatusInsufficientStorage, &ServerErrorCodeInsufficientStorage)
	// HTTP508StatusLineAndCodeLoopDetected represents 508 Loop Detected.
	HTTP508StatusLineAndCodeLoopDetected = NewHTTPStatusLineAndCode(http.StatusLoopDetected, &ServerErrorCodeLoopDetected)
	// HTTP510StatusLineAndCodeNotExtended represents 510 Not Extended.
	HTTP510StatusLineAndCodeNotExtended = NewHTTPStatusLineAndCode(http.StatusNotExtended, &ServerErrorCodeNotExtended)
	// HTTP511StatusLineAndCodeNetworkAuthenticationRequired represents 511 Network Authentication Required.
	HTTP511StatusLineAndCodeNetworkAuthenticationRequired = NewHTTPStatusLineAndCode(http.StatusNetworkAuthenticationRequired, &ServerErrorCodeNetworkAuthenticationRequired)
)

// Proprietary application error codes for HTTP status codes.
var (
	// InfoCodeContinue is the app code for 100 Continue.
	InfoCodeContinue = ProprietaryAppCode("INFO_CONTINUE")
	// InfoCodeSwitchingProtocols is the app code for 101 Switching Protocols.
	InfoCodeSwitchingProtocols = ProprietaryAppCode("INFO_SWITCHING_PROTOCOLS")
	// InfoCodeProcessing is the app code for 102 Processing.
	InfoCodeProcessing = ProprietaryAppCode("INFO_PROCESSING")
	// InfoCodeEarlyHints is the app code for 103 Early Hints.
	InfoCodeEarlyHints = ProprietaryAppCode("INFO_EARLY_HINTS")

	// SuccessCodeOK is the app code for 200 OK.
	SuccessCodeOK = ProprietaryAppCode("SUCCESS_OK")
	// SuccessCodeCreated is the app code for 201 Created.
	SuccessCodeCreated = ProprietaryAppCode("SUCCESS_CREATED")
	// SuccessCodeAccepted is the app code for 202 Accepted.
	SuccessCodeAccepted = ProprietaryAppCode("SUCCESS_ACCEPTED")
	// SuccessCodeNonAuthoritativeInfo is the app code for 203 Non-Authoritative Information.
	SuccessCodeNonAuthoritativeInfo = ProprietaryAppCode("SUCCESS_NON_AUTHORITATIVE_INFO")
	// SuccessCodeNoContent is the app code for 204 No Content.
	SuccessCodeNoContent = ProprietaryAppCode("SUCCESS_NO_CONTENT")
	// SuccessCodeResetContent is the app code for 205 Reset Content.
	SuccessCodeResetContent = ProprietaryAppCode("SUCCESS_RESET_CONTENT")
	// SuccessCodePartialContent is the app code for 206 Partial Content.
	SuccessCodePartialContent = ProprietaryAppCode("SUCCESS_PARTIAL_CONTENT")
	// SuccessCodeMultiStatus is the app code for 207 Multi-Status.
	SuccessCodeMultiStatus = ProprietaryAppCode("SUCCESS_MULTI_STATUS")
	// SuccessCodeAlreadyReported is the app code for 208 Already Reported.
	SuccessCodeAlreadyReported = ProprietaryAppCode("SUCCESS_ALREADY_REPORTED")
	// SuccessCodeIMUsed is the app code for 226 IM Used.
	SuccessCodeIMUsed = ProprietaryAppCode("SUCCESS_IM_USED")

	// RedirectionCodeMultipleChoices is the app code for 300 Multiple Choices.
	RedirectionCodeMultipleChoices = ProprietaryAppCode("REDIRECTION_MULTIPLE_CHOICES")
	// RedirectionCodeMovedPermanently is the app code for 301 Moved Permanently.
	RedirectionCodeMovedPermanently = ProprietaryAppCode("REDIRECTION_MOVED_PERMANENTLY")
	// RedirectionCodeFound is the app code for 302 Found.
	RedirectionCodeFound = ProprietaryAppCode("REDIRECTION_FOUND")
	// RedirectionCodeSeeOther is the app code for 303 See Other.
	RedirectionCodeSeeOther = ProprietaryAppCode("REDIRECTION_SEE_OTHER")
	// RedirectionCodeNotModified is the app code for 304 Not Modified.
	RedirectionCodeNotModified = ProprietaryAppCode("REDIRECTION_NOT_MODIFIED")
	// RedirectionCodeUseProxy is the app code for 305 Use Proxy.
	RedirectionCodeUseProxy = ProprietaryAppCode("REDIRECTION_USE_PROXY")
	// RedirectionCodeTemporaryRedirect is the app code for 307 Temporary Redirect.
	RedirectionCodeTemporaryRedirect = ProprietaryAppCode("REDIRECTION_TEMPORARY_REDIRECT")
	// RedirectionCodePermanentRedirect is the app code for 308 Permanent Redirect.
	RedirectionCodePermanentRedirect = ProprietaryAppCode("REDIRECTION_PERMANENT_REDIRECT")

	// ClientErrorCodeBadRequest is the app code for 400 Bad Request.
	ClientErrorCodeBadRequest = ProprietaryAppCode("CLIENT_ERROR_BAD_REQUEST")
	// ClientErrorCodeUnauthorized is the app code for 401 Unauthorized.
	ClientErrorCodeUnauthorized = ProprietaryAppCode("CLIENT_ERROR_UNAUTHORIZED")
	// ClientErrorCodePaymentRequired is the app code for 402 Payment Required.
	ClientErrorCodePaymentRequired = ProprietaryAppCode("CLIENT_ERROR_PAYMENT_REQUIRED")
	// ClientErrorCodeForbidden is the app code for 403 Forbidden.
	ClientErrorCodeForbidden = ProprietaryAppCode("CLIENT_ERROR_FORBIDDEN")
	// ClientErrorCodeNotFound is the app code for 404 Not Found.
	ClientErrorCodeNotFound = ProprietaryAppCode("CLIENT_ERROR_NOT_FOUND")
	// ClientErrorCodeMethodNotAllowed is the app code for 405 Method Not Allowed.
	ClientErrorCodeMethodNotAllowed = ProprietaryAppCode("CLIENT_ERROR_METHOD_NOT_ALLOWED")
	// ClientErrorCodeNotAcceptable is the app code for 406 Not Acceptable.
	ClientErrorCodeNotAcceptable = ProprietaryAppCode("CLIENT_ERROR_NOT_ACCEPTABLE")
	// ClientErrorCodeProxyAuthRequired is the app code for 407 Proxy Authentication Required.
	ClientErrorCodeProxyAuthRequired = ProprietaryAppCode("CLIENT_ERROR_PROXY_AUTH_REQUIRED")
	// ClientErrorCodeRequestTimeout is the app code for 408 Request Timeout.
	ClientErrorCodeRequestTimeout = ProprietaryAppCode("CLIENT_ERROR_REQUEST_TIMEOUT")
	// ClientErrorCodeConflict is the app code for 409 Conflict.
	ClientErrorCodeConflict = ProprietaryAppCode("CLIENT_ERROR_CONFLICT")
	// ClientErrorCodeGone is the app code for 410 Gone.
	ClientErrorCodeGone = ProprietaryAppCode("CLIENT_ERROR_GONE")
	// ClientErrorCodeLengthRequired is the app code for 411 Length Required.
	ClientErrorCodeLengthRequired = ProprietaryAppCode("CLIENT_ERROR_LENGTH_REQUIRED")
	// ClientErrorCodePreconditionFailed is the app code for 412 Precondition Failed.
	ClientErrorCodePreconditionFailed = ProprietaryAppCode("CLIENT_ERROR_PRECONDITION_FAILED")
	// ClientErrorCodePayloadTooLarge is the app code for 413 Payload Too Large.
	ClientErrorCodePayloadTooLarge = ProprietaryAppCode("CLIENT_ERROR_PAYLOAD_TOO_LARGE")
	// ClientErrorCodeURITooLong is the app code for 414 URI Too Long.
	ClientErrorCodeURITooLong = ProprietaryAppCode("CLIENT_ERROR_URI_TOO_LONG")
	// ClientErrorCodeUnsupportedMediaType is the app code for 415 Unsupported Media Type.
	ClientErrorCodeUnsupportedMediaType = ProprietaryAppCode("CLIENT_ERROR_UNSUPPORTED_MEDIA_TYPE")
	// ClientErrorCodeRangeNotSatisfiable is the app code for 416 Range Not Satisfiable.
	ClientErrorCodeRangeNotSatisfiable = ProprietaryAppCode("CLIENT_ERROR_RANGE_NOT_SATISFIABLE")
	// ClientErrorCodeExpectationFailed is the app code for 417 Expectation Failed.
	ClientErrorCodeExpectationFailed = ProprietaryAppCode("CLIENT_ERROR_EXPECTATION_FAILED")
	// ClientErrorCodeTeapot is the app code for 418 I'm a teapot.
	ClientErrorCodeTeapot = ProprietaryAppCode("CLIENT_ERROR_TEAPOT")
	// ClientErrorCodeMisdirectedRequest is the app code for 421 Misdirected Request.
	ClientErrorCodeMisdirectedRequest = ProprietaryAppCode("CLIENT_ERROR_MISDIRECTED_REQUEST")
	// ClientErrorCodeUnprocessableEntity is the app code for 422 Unprocessable Entity.
	ClientErrorCodeUnprocessableEntity = ProprietaryAppCode("CLIENT_ERROR_UNPROCESSABLE_ENTITY")
	// ClientErrorCodeLocked is the app code for 423 Locked.
	ClientErrorCodeLocked = ProprietaryAppCode("CLIENT_ERROR_LOCKED")
	// ClientErrorCodeFailedDependency is the app code for 424 Failed Dependency.
	ClientErrorCodeFailedDependency = ProprietaryAppCode("CLIENT_ERROR_FAILED_DEPENDENCY")
	// ClientErrorCodeTooEarly is the app code for 425 Too Early.
	ClientErrorCodeTooEarly = ProprietaryAppCode("CLIENT_ERROR_TOO_EARLY")
	// ClientErrorCodeUpgradeRequired is the app code for 426 Upgrade Required.
	ClientErrorCodeUpgradeRequired = ProprietaryAppCode("CLIENT_ERROR_UPGRADE_REQUIRED")
	// ClientErrorCodePreconditionRequired is the app code for 428 Precondition Required.
	ClientErrorCodePreconditionRequired = ProprietaryAppCode("CLIENT_ERROR_PRECONDITION_REQUIRED")
	// ClientErrorCodeTooManyRequests is the app code for 429 Too Many Requests.
	ClientErrorCodeTooManyRequests = ProprietaryAppCode("CLIENT_ERROR_TOO_MANY_REQUESTS")
	// ClientErrorCodeRequestHeaderFieldsTooLarge is the app code for 431 Request Header Fields Too Large.
	ClientErrorCodeRequestHeaderFieldsTooLarge = ProprietaryAppCode("CLIENT_ERROR_REQUEST_HEADER_FIELDS_TOO_LARGE")
	// ClientErrorCodeUnavailableForLegalReasons is the app code for 451 Unavailable For Legal Reasons.
	ClientErrorCodeUnavailableForLegalReasons = ProprietaryAppCode("CLIENT_ERROR_UNAVAILABLE_FOR_LEGAL_REASONS")

	// ServerErrorCodeInternalServerServerError is the app code for 500 Internal Server Error.
	ServerErrorCodeInternalServerServerError = ProprietaryAppCode("SERVER_ERROR_INTERNAL_SERVER_ERROR")
	// ServerErrorCodeNotImplemented is the app code for 501 Not Implemented.
	ServerErrorCodeNotImplemented = ProprietaryAppCode("SERVER_ERROR_NOT_IMPLEMENTED")
	// ServerErrorCodeBadGateway is the app code for 502 Bad Gateway.
	ServerErrorCodeBadGateway = ProprietaryAppCode("SERVER_ERROR_BAD_GATEWAY")
	// ServerErrorCodeServiceUnavailable is the app code for 503 Service Unavailable.
	ServerErrorCodeServiceUnavailable = ProprietaryAppCode("SERVER_ERROR_SERVICE_UNAVAILABLE")
	// ServerErrorCodeGatewayTimeout is the app code for 504 Gateway Timeout.
	ServerErrorCodeGatewayTimeout = ProprietaryAppCode("SERVER_ERROR_GATEWAY_TIMEOUT")
	// ServerErrorCodeHTTPVersionNotSupported is the app code for 505 HTTP Version Not Supported.
	ServerErrorCodeHTTPVersionNotSupported = ProprietaryAppCode("SERVER_ERROR_HTTP_VERSION_NOT_SUPPORTED")
	// ServerErrorCodeVariantAlsoNegotiates is the app code for 506 Variant Also Negotiates.
	ServerErrorCodeVariantAlsoNegotiates = ProprietaryAppCode("SERVER_ERROR_VARIANT_ALSO_NEGOTIATES")
	// ServerErrorCodeInsufficientStorage is the app code for 507 Insufficient Storage.
	ServerErrorCodeInsufficientStorage = ProprietaryAppCode("SERVER_ERROR_INSUFFICIENT_STORAGE")
	// ServerErrorCodeLoopDetected is the app code for 508 Loop Detected.
	ServerErrorCodeLoopDetected = ProprietaryAppCode("SERVER_ERROR_LOOP_DETECTED")
	// ServerErrorCodeNotExtended is the app code for 510 Not Extended.
	ServerErrorCodeNotExtended = ProprietaryAppCode("SERVER_ERROR_NOT_EXTENDED")
	// ServerErrorCodeNetworkAuthenticationRequired is the app code for 511 Network Authentication Required.
	ServerErrorCodeNetworkAuthenticationRequired = ProprietaryAppCode("SERVER_ERROR_NETWORK_AUTHENTICATION_REQUIRED")
)
