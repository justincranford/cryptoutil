package apperr

import "net/http"

type (
	HTTPStatusCode   int
	HTTPReasonPhrase string
	Code             string
)

type HTTPStatusLine struct {
	StatusCode   HTTPStatusCode   // HTTP status code
	ReasonPhrase HTTPReasonPhrase // HTTP reason phrase
}

type HTTPStatusLineAndCode struct {
	StatusLine HTTPStatusLine // HTTP status line (HTTP status code + HTTP reason phrase)
	Code       Code           // Code
}

func NewHTTPStatusLineAndCode(statusCode HTTPStatusCode, code Code) (HTTPStatusLine, HTTPStatusLineAndCode) {
	httpStatusLine := NewHTTPStatusLine(statusCode, HTTPReasonPhrase(http.StatusText(int(statusCode))))
	return httpStatusLine, HTTPStatusLineAndCode{StatusLine: httpStatusLine, Code: code}
}

func NewHTTPStatusLine(statusCode HTTPStatusCode, reasonPhrase HTTPReasonPhrase) HTTPStatusLine {
	return HTTPStatusLine{StatusCode: statusCode, ReasonPhrase: reasonPhrase}
}

func NewCode(message string) Code {
	return Code(message)
}

// Generate all HTTP status codes (100-599) using net/http constants
var (
	// 1xx Informational
	HTTP100StatusLineContinue, HTTP100StatusLineAndCodeContinue                     = NewHTTPStatusLineAndCode(http.StatusContinue, InfoCodeContinue)
	HTTP101StatusLineSwitchingProtocols, HTTP101StatusLineAndCodeSwitchingProtocols = NewHTTPStatusLineAndCode(http.StatusSwitchingProtocols, InfoCodeSwitchingProtocols)
	HTTP102StatusLineProcessing, HTTP102StatusLineAndCodeProcessing                 = NewHTTPStatusLineAndCode(http.StatusProcessing, InfoCodeProcessing)
	HTTP103StatusLineEarlyHints, HTTP103StatusLineAndCodeEarlyHints                 = NewHTTPStatusLineAndCode(http.StatusEarlyHints, InfoCodeEarlyHints)

	// 2xx Success
	HTTP200StatusLineOK, HTTP200StatusLineAndCodeOK                                     = NewHTTPStatusLineAndCode(http.StatusOK, SuccessCodeOK)
	HTTP201StatusLineCreated, HTTP201StatusLineAndCodeCreated                           = NewHTTPStatusLineAndCode(http.StatusCreated, SuccessCodeCreated)
	HTTP202StatusLineAccepted, HTTP202StatusLineAndCodeAccepted                         = NewHTTPStatusLineAndCode(http.StatusAccepted, SuccessCodeAccepted)
	HTTP203StatusLineNonAuthoritativeInfo, HTTP203StatusLineAndCodeNonAuthoritativeInfo = NewHTTPStatusLineAndCode(http.StatusNonAuthoritativeInfo, SuccessCodeNonAuthoritativeInfo)
	HTTP204StatusLineNoContent, HTTP204StatusLineAndCodeNoContent                       = NewHTTPStatusLineAndCode(http.StatusNoContent, SuccessCodeNoContent)
	HTTP205StatusLineResetContent, HTTP205StatusLineAndCodeResetContent                 = NewHTTPStatusLineAndCode(http.StatusResetContent, SuccessCodeResetContent)
	HTTP206StatusLinePartialContent, HTTP206StatusLineAndCodePartialContent             = NewHTTPStatusLineAndCode(http.StatusPartialContent, SuccessCodePartialContent)
	HTTP207StatusLineMultiStatus, HTTP207StatusLineAndCodeMultiStatus                   = NewHTTPStatusLineAndCode(http.StatusMultiStatus, SuccessCodeMultiStatus)
	HTTP208StatusLineAlreadyReported, HTTP208StatusLineAndCodeAlreadyReported           = NewHTTPStatusLineAndCode(http.StatusAlreadyReported, SuccessCodeAlreadyReported)
	HTTP226StatusLineIMUsed, HTTP226StatusLineAndCodeIMUsed                             = NewHTTPStatusLineAndCode(http.StatusIMUsed, SuccessCodeIMUsed)

	// 3xx Redirection
	HTTP300StatusLineMultipleChoices, HTTP300StatusLineAndCodeMultipleChoices     = NewHTTPStatusLineAndCode(http.StatusMultipleChoices, RedirectionCodeMultipleChoices)
	HTTP301StatusLineMovedPermanently, HTTP301StatusLineAndCodeMovedPermanently   = NewHTTPStatusLineAndCode(http.StatusMovedPermanently, RedirectionCodeMovedPermanently)
	HTTP302StatusLineFound, HTTP302StatusLineAndCodeFound                         = NewHTTPStatusLineAndCode(http.StatusFound, RedirectionCodeFound)
	HTTP303StatusLineSeeOther, HTTP303StatusLineAndCodeSeeOther                   = NewHTTPStatusLineAndCode(http.StatusSeeOther, RedirectionCodeSeeOther)
	HTTP304StatusLineNotModified, HTTP304StatusLineAndCodeNotModified             = NewHTTPStatusLineAndCode(http.StatusNotModified, RedirectionCodeNotModified)
	HTTP305StatusLineUseProxy, HTTP305StatusLineAndCodeUseProxy                   = NewHTTPStatusLineAndCode(http.StatusUseProxy, RedirectionCodeUseProxy)
	HTTP307StatusLineTemporaryRedirect, HTTP307StatusLineAndCodeTemporaryRedirect = NewHTTPStatusLineAndCode(http.StatusTemporaryRedirect, RedirectionCodeTemporaryRedirect)
	HTTP308StatusLinePermanentRedirect, HTTP308StatusLineAndCodePermanentRedirect = NewHTTPStatusLineAndCode(http.StatusPermanentRedirect, RedirectionCodePermanentRedirect)

	// 4xx Client Errors
	HTTP400StatusLineBadRequest, HTTP400StatusLineAndCodeBadRequest                                   = NewHTTPStatusLineAndCode(http.StatusBadRequest, ClientErrorCodeBadRequest)
	HTTP401StatusLineUnauthorized, HTTP401StatusLineAndCodeUnauthorized                               = NewHTTPStatusLineAndCode(http.StatusUnauthorized, ClientErrorCodeUnauthorized)
	HTTP402StatusLinePaymentRequired, HTTP402StatusLineAndCodePaymentRequired                         = NewHTTPStatusLineAndCode(http.StatusPaymentRequired, ClientErrorCodePaymentRequired)
	HTTP403StatusLineForbidden, HTTP403StatusLineAndCodeForbidden                                     = NewHTTPStatusLineAndCode(http.StatusForbidden, ClientErrorCodeForbidden)
	HTTP404StatusLineNotFound, HTTP404StatusLineAndCodeNotFound                                       = NewHTTPStatusLineAndCode(http.StatusNotFound, ClientErrorCodeNotFound)
	HTTP405StatusLineMethodNotAllowed, HTTP405StatusLineAndCodeMethodNotAllowed                       = NewHTTPStatusLineAndCode(http.StatusMethodNotAllowed, ClientErrorCodeMethodNotAllowed)
	HTTP406StatusLineNotAcceptable, HTTP406StatusLineAndCodeNotAcceptable                             = NewHTTPStatusLineAndCode(http.StatusNotAcceptable, ClientErrorCodeNotAcceptable)
	HTTP407StatusLineProxyAuthRequired, HTTP407StatusLineAndCodeProxyAuthRequired                     = NewHTTPStatusLineAndCode(http.StatusProxyAuthRequired, ClientErrorCodeProxyAuthRequired)
	HTTP408StatusLineRequestTimeout, HTTP408StatusLineAndCodeRequestTimeout                           = NewHTTPStatusLineAndCode(http.StatusRequestTimeout, ClientErrorCodeRequestTimeout)
	HTTP409StatusLineConflict, HTTP409StatusLineAndCodeConflict                                       = NewHTTPStatusLineAndCode(http.StatusConflict, ClientErrorCodeConflict)
	HTTP410StatusLineGone, HTTP410StatusLineAndCodeGone                                               = NewHTTPStatusLineAndCode(http.StatusGone, ClientErrorCodeGone)
	HTTP411StatusLineLengthRequired, HTTP411StatusLineAndCodeLengthRequired                           = NewHTTPStatusLineAndCode(http.StatusLengthRequired, ClientErrorCodeLengthRequired)
	HTTP412StatusLinePreconditionFailed, HTTP412StatusLineAndCodePreconditionFailed                   = NewHTTPStatusLineAndCode(http.StatusPreconditionFailed, ClientErrorCodePreconditionFailed)
	HTTP413StatusLinePayloadTooLarge, HTTP413StatusLineAndCodePayloadTooLarge                         = NewHTTPStatusLineAndCode(http.StatusRequestEntityTooLarge, ClientErrorCodePayloadTooLarge)
	HTTP414StatusLineURITooLong, HTTP414StatusLineAndCodeURITooLong                                   = NewHTTPStatusLineAndCode(http.StatusRequestURITooLong, ClientErrorCodeURITooLong)
	HTTP415StatusLineUnsupportedMediaType, HTTP415StatusLineAndCodeUnsupportedMediaType               = NewHTTPStatusLineAndCode(http.StatusUnsupportedMediaType, ClientErrorCodeUnsupportedMediaType)
	HTTP416StatusLineRangeNotSatisfiable, HTTP416StatusLineAndCodeRangeNotSatisfiable                 = NewHTTPStatusLineAndCode(http.StatusRequestedRangeNotSatisfiable, ClientErrorCodeRangeNotSatisfiable)
	HTTP417StatusLineExpectationFailed, HTTP417StatusLineAndCodeExpectationFailed                     = NewHTTPStatusLineAndCode(http.StatusExpectationFailed, ClientErrorCodeExpectationFailed)
	HTTP418StatusLineTeapot, HTTP418StatusLineAndCodeTeapot                                           = NewHTTPStatusLineAndCode(http.StatusTeapot, ClientErrorCodeTeapot)
	HTTP421StatusLineMisdirectedRequest, HTTP421StatusLineAndCodeMisdirectedRequest                   = NewHTTPStatusLineAndCode(http.StatusMisdirectedRequest, ClientErrorCodeMisdirectedRequest)
	HTTP422StatusLineUnprocessableEntity, HTTP422StatusLineAndCodeUnprocessableEntity                 = NewHTTPStatusLineAndCode(http.StatusUnprocessableEntity, ClientErrorCodeUnprocessableEntity)
	HTTP423StatusLineLocked, HTTP423StatusLineAndCodeLocked                                           = NewHTTPStatusLineAndCode(http.StatusLocked, ClientErrorCodeLocked)
	HTTP424StatusLineFailedDependency, HTTP424StatusLineAndCodeFailedDependency                       = NewHTTPStatusLineAndCode(http.StatusFailedDependency, ClientErrorCodeFailedDependency)
	HTTP425StatusLineTooEarly, HTTP425StatusLineAndCodeTooEarly                                       = NewHTTPStatusLineAndCode(http.StatusTooEarly, ClientErrorCodeTooEarly)
	HTTP426StatusLineUpgradeRequired, HTTP426StatusLineAndCodeUpgradeRequired                         = NewHTTPStatusLineAndCode(http.StatusUpgradeRequired, ClientErrorCodeUpgradeRequired)
	HTTP428StatusLinePreconditionRequired, HTTP428StatusLineAndCodePreconditionRequired               = NewHTTPStatusLineAndCode(http.StatusPreconditionRequired, ClientErrorCodePreconditionRequired)
	HTTP429StatusLineTooManyRequests, HTTP429StatusLineAndCodeTooManyRequests                         = NewHTTPStatusLineAndCode(http.StatusTooManyRequests, ClientErrorCodeTooManyRequests)
	HTTP431StatusLineRequestHeaderFieldsTooLarge, HTTP431StatusLineAndCodeRequestHeaderFieldsTooLarge = NewHTTPStatusLineAndCode(http.StatusRequestHeaderFieldsTooLarge, ClientErrorCodeRequestHeaderFieldsTooLarge)
	HTTP451StatusLineUnavailableForLegalReasons, HTTP451StatusLineAndCodeUnavailableForLegalReasons   = NewHTTPStatusLineAndCode(http.StatusUnavailableForLegalReasons, ClientErrorCodeUnavailableForLegalReasons)

	// 5xx Server Errors
	HTTP500StatusLineInternalServerError, HTTP500StatusLineAndCodeInternalServerError                     = NewHTTPStatusLineAndCode(http.StatusInternalServerError, ServerErrorCodeInternalServerServerError)
	HTTP501StatusLineNotImplemented, HTTP501StatusLineAndCodeNotImplemented                               = NewHTTPStatusLineAndCode(http.StatusNotImplemented, ServerErrorCodeNotImplemented)
	HTTP502StatusLineBadGateway, HTTP502StatusLineAndCodeBadGateway                                       = NewHTTPStatusLineAndCode(http.StatusBadGateway, ServerErrorCodeBadGateway)
	HTTP503StatusLineServiceUnavailable, HTTP503StatusLineAndCodeServiceUnavailable                       = NewHTTPStatusLineAndCode(http.StatusServiceUnavailable, ServerErrorCodeServiceUnavailable)
	HTTP504StatusLineGatewayTimeout, HTTP504StatusLineAndCodeGatewayTimeout                               = NewHTTPStatusLineAndCode(http.StatusGatewayTimeout, ServerErrorCodeGatewayTimeout)
	HTTP505StatusLineHTTPVersionNotSupported, HTTP505StatusLineAndCodeHTTPVersionNotSupported             = NewHTTPStatusLineAndCode(http.StatusHTTPVersionNotSupported, ServerErrorCodeHTTPVersionNotSupported)
	HTTP506StatusLineVariantAlsoNegotiates, HTTP506StatusLineAndCodeVariantAlsoNegotiates                 = NewHTTPStatusLineAndCode(http.StatusVariantAlsoNegotiates, ServerErrorCodeVariantAlsoNegotiates)
	HTTP507StatusLineInsufficientStorage, HTTP507StatusLineAndCodeInsufficientStorage                     = NewHTTPStatusLineAndCode(http.StatusInsufficientStorage, ServerErrorCodeInsufficientStorage)
	HTTP508StatusLineLoopDetected, HTTP508StatusLineAndCodeLoopDetected                                   = NewHTTPStatusLineAndCode(http.StatusLoopDetected, ServerErrorCodeLoopDetected)
	HTTP510StatusLineNotExtended, HTTP510StatusLineAndCodeNotExtended                                     = NewHTTPStatusLineAndCode(http.StatusNotExtended, ServerErrorCodeNotExtended)
	HTTP511StatusLineNetworkAuthenticationRequired, HTTP511StatusLineAndCodeNetworkAuthenticationRequired = NewHTTPStatusLineAndCode(http.StatusNetworkAuthenticationRequired, ServerErrorCodeNetworkAuthenticationRequired)
)

var (
	// 1xx Informational
	InfoCodeContinue           = Code("INFO_CONTINUE")
	InfoCodeSwitchingProtocols = Code("INFO_SWITCHING_PROTOCOLS")
	InfoCodeProcessing         = Code("INFO_PROCESSING")
	InfoCodeEarlyHints         = Code("INFO_EARLY_HINTS")

	// 2xx Success
	SuccessCodeOK                   = Code("SUCCESS_OK")
	SuccessCodeCreated              = Code("SUCCESS_CREATED")
	SuccessCodeAccepted             = Code("SUCCESS_ACCEPTED")
	SuccessCodeNonAuthoritativeInfo = Code("SUCCESS_NON_AUTHORITATIVE_INFO")
	SuccessCodeNoContent            = Code("SUCCESS_NO_CONTENT")
	SuccessCodeResetContent         = Code("SUCCESS_RESET_CONTENT")
	SuccessCodePartialContent       = Code("SUCCESS_PARTIAL_CONTENT")
	SuccessCodeMultiStatus          = Code("SUCCESS_MULTI_STATUS")
	SuccessCodeAlreadyReported      = Code("SUCCESS_ALREADY_REPORTED")
	SuccessCodeIMUsed               = Code("SUCCESS_IM_USED")

	// 3xx Redirection
	RedirectionCodeMultipleChoices   = Code("REDIRECTION_MULTIPLE_CHOICES")
	RedirectionCodeMovedPermanently  = Code("REDIRECTION_MOVED_PERMANENTLY")
	RedirectionCodeFound             = Code("REDIRECTION_FOUND")
	RedirectionCodeSeeOther          = Code("REDIRECTION_SEE_OTHER")
	RedirectionCodeNotModified       = Code("REDIRECTION_NOT_MODIFIED")
	RedirectionCodeUseProxy          = Code("REDIRECTION_USE_PROXY")
	RedirectionCodeTemporaryRedirect = Code("REDIRECTION_TEMPORARY_REDIRECT")
	RedirectionCodePermanentRedirect = Code("REDIRECTION_PERMANENT_REDIRECT")

	// 4xx Client Errors
	ClientErrorCodeBadRequest                  = Code("CLIENT_ERROR_BAD_REQUEST")
	ClientErrorCodeUnauthorized                = Code("CLIENT_ERROR_UNAUTHORIZED")
	ClientErrorCodePaymentRequired             = Code("CLIENT_ERROR_PAYMENT_REQUIRED")
	ClientErrorCodeForbidden                   = Code("CLIENT_ERROR_FORBIDDEN")
	ClientErrorCodeNotFound                    = Code("CLIENT_ERROR_NOT_FOUND")
	ClientErrorCodeMethodNotAllowed            = Code("CLIENT_ERROR_METHOD_NOT_ALLOWED")
	ClientErrorCodeNotAcceptable               = Code("CLIENT_ERROR_NOT_ACCEPTABLE")
	ClientErrorCodeProxyAuthRequired           = Code("CLIENT_ERROR_PROXY_AUTH_REQUIRED")
	ClientErrorCodeRequestTimeout              = Code("CLIENT_ERROR_REQUEST_TIMEOUT")
	ClientErrorCodeConflict                    = Code("CLIENT_ERROR_CONFLICT")
	ClientErrorCodeGone                        = Code("CLIENT_ERROR_GONE")
	ClientErrorCodeLengthRequired              = Code("CLIENT_ERROR_LENGTH_REQUIRED")
	ClientErrorCodePreconditionFailed          = Code("CLIENT_ERROR_PRECONDITION_FAILED")
	ClientErrorCodePayloadTooLarge             = Code("CLIENT_ERROR_PAYLOAD_TOO_LARGE")
	ClientErrorCodeURITooLong                  = Code("CLIENT_ERROR_URI_TOO_LONG")
	ClientErrorCodeUnsupportedMediaType        = Code("CLIENT_ERROR_UNSUPPORTED_MEDIA_TYPE")
	ClientErrorCodeRangeNotSatisfiable         = Code("CLIENT_ERROR_RANGE_NOT_SATISFIABLE")
	ClientErrorCodeExpectationFailed           = Code("CLIENT_ERROR_EXPECTATION_FAILED")
	ClientErrorCodeTeapot                      = Code("CLIENT_ERROR_TEAPOT")
	ClientErrorCodeMisdirectedRequest          = Code("CLIENT_ERROR_MISDIRECTED_REQUEST")
	ClientErrorCodeUnprocessableEntity         = Code("CLIENT_ERROR_UNPROCESSABLE_ENTITY")
	ClientErrorCodeLocked                      = Code("CLIENT_ERROR_LOCKED")
	ClientErrorCodeFailedDependency            = Code("CLIENT_ERROR_FAILED_DEPENDENCY")
	ClientErrorCodeTooEarly                    = Code("CLIENT_ERROR_TOO_EARLY")
	ClientErrorCodeUpgradeRequired             = Code("CLIENT_ERROR_UPGRADE_REQUIRED")
	ClientErrorCodePreconditionRequired        = Code("CLIENT_ERROR_PRECONDITION_REQUIRED")
	ClientErrorCodeTooManyRequests             = Code("CLIENT_ERROR_TOO_MANY_REQUESTS")
	ClientErrorCodeRequestHeaderFieldsTooLarge = Code("CLIENT_ERROR_REQUEST_HEADER_FIELDS_TOO_LARGE")
	ClientErrorCodeUnavailableForLegalReasons  = Code("CLIENT_ERROR_UNAVAILABLE_FOR_LEGAL_REASONS")

	// 5xx Server Errors
	ServerErrorCodeInternalServerServerError     = Code("SERVER_ERROR_INTERNAL_SERVER_ERROR")
	ServerErrorCodeNotImplemented                = Code("SERVER_ERROR_NOT_IMPLEMENTED")
	ServerErrorCodeBadGateway                    = Code("SERVER_ERROR_BAD_GATEWAY")
	ServerErrorCodeServiceUnavailable            = Code("SERVER_ERROR_SERVICE_UNAVAILABLE")
	ServerErrorCodeGatewayTimeout                = Code("SERVER_ERROR_GATEWAY_TIMEOUT")
	ServerErrorCodeHTTPVersionNotSupported       = Code("SERVER_ERROR_HTTP_VERSION_NOT_SUPPORTED")
	ServerErrorCodeVariantAlsoNegotiates         = Code("SERVER_ERROR_VARIANT_ALSO_NEGOTIATES")
	ServerErrorCodeInsufficientStorage           = Code("SERVER_ERROR_INSUFFICIENT_STORAGE")
	ServerErrorCodeLoopDetected                  = Code("SERVER_ERROR_LOOP_DETECTED")
	ServerErrorCodeNotExtended                   = Code("SERVER_ERROR_NOT_EXTENDED")
	ServerErrorCodeNetworkAuthenticationRequired = Code("SERVER_ERROR_NETWORK_AUTHENTICATION_REQUIRED")
)
