package apperr

import "net/http"

type (
	HTTPStatusCode     int
	HTTPReasonPhrase   string
	ProprietaryAppCode string
)

type HTTPStatusLine struct {
	StatusCode   HTTPStatusCode   // HTTP status code
	ReasonPhrase HTTPReasonPhrase // HTTP reason phrase
}

type HTTPStatusLineAndCode struct {
	StatusLine HTTPStatusLine     // HTTP status line (HTTP status code + HTTP reason phrase)
	Code       ProprietaryAppCode // Proprietary Application Code
}

func NewHTTPStatusLineAndCode(httpStatusCode HTTPStatusCode, proprietartAppCode *ProprietaryAppCode) HTTPStatusLineAndCode {
	return HTTPStatusLineAndCode{
		StatusLine: HTTPStatusLine{
			StatusCode:   httpStatusCode,                                         // 100-599
			ReasonPhrase: HTTPReasonPhrase(http.StatusText(int(httpStatusCode))), // standard HTTP reason phrase from standard HTTP status codes
		},
		Code: *proprietartAppCode, // Proprietary application code
	}
}

func NewHTTPStatusLine(statusCode HTTPStatusCode, reasonPhrase HTTPReasonPhrase) HTTPStatusLine {
	return HTTPStatusLine{StatusCode: statusCode, ReasonPhrase: reasonPhrase}
}

func NewCode(message string) ProprietaryAppCode {
	return ProprietaryAppCode(message)
}

// Generate all HTTP status codes (100-599) using net/http constants
var (
	// 1xx Informational
	HTTP100StatusLineAndCodeContinue           = NewHTTPStatusLineAndCode(http.StatusContinue, &InfoCodeContinue)
	HTTP101StatusLineAndCodeSwitchingProtocols = NewHTTPStatusLineAndCode(http.StatusSwitchingProtocols, &InfoCodeSwitchingProtocols)
	HTTP102StatusLineAndCodeProcessing         = NewHTTPStatusLineAndCode(http.StatusProcessing, &InfoCodeProcessing)
	HTTP103StatusLineAndCodeEarlyHints         = NewHTTPStatusLineAndCode(http.StatusEarlyHints, &InfoCodeEarlyHints)

	// 2xx Success
	HTTP200StatusLineAndCodeOK                   = NewHTTPStatusLineAndCode(http.StatusOK, &SuccessCodeOK)
	HTTP201StatusLineAndCodeCreated              = NewHTTPStatusLineAndCode(http.StatusCreated, &SuccessCodeCreated)
	HTTP202StatusLineAndCodeAccepted             = NewHTTPStatusLineAndCode(http.StatusAccepted, &SuccessCodeAccepted)
	HTTP203StatusLineAndCodeNonAuthoritativeInfo = NewHTTPStatusLineAndCode(http.StatusNonAuthoritativeInfo, &SuccessCodeNonAuthoritativeInfo)
	HTTP204StatusLineAndCodeNoContent            = NewHTTPStatusLineAndCode(http.StatusNoContent, &SuccessCodeNoContent)
	HTTP205StatusLineAndCodeResetContent         = NewHTTPStatusLineAndCode(http.StatusResetContent, &SuccessCodeResetContent)
	HTTP206StatusLineAndCodePartialContent       = NewHTTPStatusLineAndCode(http.StatusPartialContent, &SuccessCodePartialContent)
	HTTP207StatusLineAndCodeMultiStatus          = NewHTTPStatusLineAndCode(http.StatusMultiStatus, &SuccessCodeMultiStatus)
	HTTP208StatusLineAndCodeAlreadyReported      = NewHTTPStatusLineAndCode(http.StatusAlreadyReported, &SuccessCodeAlreadyReported)
	HTTP226StatusLineAndCodeIMUsed               = NewHTTPStatusLineAndCode(http.StatusIMUsed, &SuccessCodeIMUsed)

	// 3xx Redirection
	HTTP300StatusLineAndCodeMultipleChoices   = NewHTTPStatusLineAndCode(http.StatusMultipleChoices, &RedirectionCodeMultipleChoices)
	HTTP301StatusLineAndCodeMovedPermanently  = NewHTTPStatusLineAndCode(http.StatusMovedPermanently, &RedirectionCodeMovedPermanently)
	HTTP302StatusLineAndCodeFound             = NewHTTPStatusLineAndCode(http.StatusFound, &RedirectionCodeFound)
	HTTP303StatusLineAndCodeSeeOther          = NewHTTPStatusLineAndCode(http.StatusSeeOther, &RedirectionCodeSeeOther)
	HTTP304StatusLineAndCodeNotModified       = NewHTTPStatusLineAndCode(http.StatusNotModified, &RedirectionCodeNotModified)
	HTTP305StatusLineAndCodeUseProxy          = NewHTTPStatusLineAndCode(http.StatusUseProxy, &RedirectionCodeUseProxy)
	HTTP307StatusLineAndCodeTemporaryRedirect = NewHTTPStatusLineAndCode(http.StatusTemporaryRedirect, &RedirectionCodeTemporaryRedirect)
	HTTP308StatusLineAndCodePermanentRedirect = NewHTTPStatusLineAndCode(http.StatusPermanentRedirect, &RedirectionCodePermanentRedirect)

	// 4xx Client Errors
	HTTP400StatusLineAndCodeBadRequest                  = NewHTTPStatusLineAndCode(http.StatusBadRequest, &ClientErrorCodeBadRequest)
	HTTP401StatusLineAndCodeUnauthorized                = NewHTTPStatusLineAndCode(http.StatusUnauthorized, &ClientErrorCodeUnauthorized)
	HTTP402StatusLineAndCodePaymentRequired             = NewHTTPStatusLineAndCode(http.StatusPaymentRequired, &ClientErrorCodePaymentRequired)
	HTTP403StatusLineAndCodeForbidden                   = NewHTTPStatusLineAndCode(http.StatusForbidden, &ClientErrorCodeForbidden)
	HTTP404StatusLineAndCodeNotFound                    = NewHTTPStatusLineAndCode(http.StatusNotFound, &ClientErrorCodeNotFound)
	HTTP405StatusLineAndCodeMethodNotAllowed            = NewHTTPStatusLineAndCode(http.StatusMethodNotAllowed, &ClientErrorCodeMethodNotAllowed)
	HTTP406StatusLineAndCodeNotAcceptable               = NewHTTPStatusLineAndCode(http.StatusNotAcceptable, &ClientErrorCodeNotAcceptable)
	HTTP407StatusLineAndCodeProxyAuthRequired           = NewHTTPStatusLineAndCode(http.StatusProxyAuthRequired, &ClientErrorCodeProxyAuthRequired)
	HTTP408StatusLineAndCodeRequestTimeout              = NewHTTPStatusLineAndCode(http.StatusRequestTimeout, &ClientErrorCodeRequestTimeout)
	HTTP409StatusLineAndCodeConflict                    = NewHTTPStatusLineAndCode(http.StatusConflict, &ClientErrorCodeConflict)
	HTTP410StatusLineAndCodeGone                        = NewHTTPStatusLineAndCode(http.StatusGone, &ClientErrorCodeGone)
	HTTP411StatusLineAndCodeLengthRequired              = NewHTTPStatusLineAndCode(http.StatusLengthRequired, &ClientErrorCodeLengthRequired)
	HTTP412StatusLineAndCodePreconditionFailed          = NewHTTPStatusLineAndCode(http.StatusPreconditionFailed, &ClientErrorCodePreconditionFailed)
	HTTP413StatusLineAndCodePayloadTooLarge             = NewHTTPStatusLineAndCode(http.StatusRequestEntityTooLarge, &ClientErrorCodePayloadTooLarge)
	HTTP414StatusLineAndCodeURITooLong                  = NewHTTPStatusLineAndCode(http.StatusRequestURITooLong, &ClientErrorCodeURITooLong)
	HTTP415StatusLineAndCodeUnsupportedMediaType        = NewHTTPStatusLineAndCode(http.StatusUnsupportedMediaType, &ClientErrorCodeUnsupportedMediaType)
	HTTP416StatusLineAndCodeRangeNotSatisfiable         = NewHTTPStatusLineAndCode(http.StatusRequestedRangeNotSatisfiable, &ClientErrorCodeRangeNotSatisfiable)
	HTTP417StatusLineAndCodeExpectationFailed           = NewHTTPStatusLineAndCode(http.StatusExpectationFailed, &ClientErrorCodeExpectationFailed)
	HTTP418StatusLineAndCodeTeapot                      = NewHTTPStatusLineAndCode(http.StatusTeapot, &ClientErrorCodeTeapot)
	HTTP421StatusLineAndCodeMisdirectedRequest          = NewHTTPStatusLineAndCode(http.StatusMisdirectedRequest, &ClientErrorCodeMisdirectedRequest)
	HTTP422StatusLineAndCodeUnprocessableEntity         = NewHTTPStatusLineAndCode(http.StatusUnprocessableEntity, &ClientErrorCodeUnprocessableEntity)
	HTTP423StatusLineAndCodeLocked                      = NewHTTPStatusLineAndCode(http.StatusLocked, &ClientErrorCodeLocked)
	HTTP424StatusLineAndCodeFailedDependency            = NewHTTPStatusLineAndCode(http.StatusFailedDependency, &ClientErrorCodeFailedDependency)
	HTTP425StatusLineAndCodeTooEarly                    = NewHTTPStatusLineAndCode(http.StatusTooEarly, &ClientErrorCodeTooEarly)
	HTTP426StatusLineAndCodeUpgradeRequired             = NewHTTPStatusLineAndCode(http.StatusUpgradeRequired, &ClientErrorCodeUpgradeRequired)
	HTTP428StatusLineAndCodePreconditionRequired        = NewHTTPStatusLineAndCode(http.StatusPreconditionRequired, &ClientErrorCodePreconditionRequired)
	HTTP429StatusLineAndCodeTooManyRequests             = NewHTTPStatusLineAndCode(http.StatusTooManyRequests, &ClientErrorCodeTooManyRequests)
	HTTP431StatusLineAndCodeRequestHeaderFieldsTooLarge = NewHTTPStatusLineAndCode(http.StatusRequestHeaderFieldsTooLarge, &ClientErrorCodeRequestHeaderFieldsTooLarge)
	HTTP451StatusLineAndCodeUnavailableForLegalReasons  = NewHTTPStatusLineAndCode(http.StatusUnavailableForLegalReasons, &ClientErrorCodeUnavailableForLegalReasons)

	// 5xx Server Errors
	HTTP500StatusLineAndCodeInternalServerError           = NewHTTPStatusLineAndCode(http.StatusInternalServerError, &ServerErrorCodeInternalServerServerError)
	HTTP501StatusLineAndCodeNotImplemented                = NewHTTPStatusLineAndCode(http.StatusNotImplemented, &ServerErrorCodeNotImplemented)
	HTTP502StatusLineAndCodeBadGateway                    = NewHTTPStatusLineAndCode(http.StatusBadGateway, &ServerErrorCodeBadGateway)
	HTTP503StatusLineAndCodeServiceUnavailable            = NewHTTPStatusLineAndCode(http.StatusServiceUnavailable, &ServerErrorCodeServiceUnavailable)
	HTTP504StatusLineAndCodeGatewayTimeout                = NewHTTPStatusLineAndCode(http.StatusGatewayTimeout, &ServerErrorCodeGatewayTimeout)
	HTTP505StatusLineAndCodeHTTPVersionNotSupported       = NewHTTPStatusLineAndCode(http.StatusHTTPVersionNotSupported, &ServerErrorCodeHTTPVersionNotSupported)
	HTTP506StatusLineAndCodeVariantAlsoNegotiates         = NewHTTPStatusLineAndCode(http.StatusVariantAlsoNegotiates, &ServerErrorCodeVariantAlsoNegotiates)
	HTTP507StatusLineAndCodeInsufficientStorage           = NewHTTPStatusLineAndCode(http.StatusInsufficientStorage, &ServerErrorCodeInsufficientStorage)
	HTTP508StatusLineAndCodeLoopDetected                  = NewHTTPStatusLineAndCode(http.StatusLoopDetected, &ServerErrorCodeLoopDetected)
	HTTP510StatusLineAndCodeNotExtended                   = NewHTTPStatusLineAndCode(http.StatusNotExtended, &ServerErrorCodeNotExtended)
	HTTP511StatusLineAndCodeNetworkAuthenticationRequired = NewHTTPStatusLineAndCode(http.StatusNetworkAuthenticationRequired, &ServerErrorCodeNetworkAuthenticationRequired)
)

var (
	// 1xx Informational
	InfoCodeContinue           = ProprietaryAppCode("INFO_CONTINUE")
	InfoCodeSwitchingProtocols = ProprietaryAppCode("INFO_SWITCHING_PROTOCOLS")
	InfoCodeProcessing         = ProprietaryAppCode("INFO_PROCESSING")
	InfoCodeEarlyHints         = ProprietaryAppCode("INFO_EARLY_HINTS")

	// 2xx Success
	SuccessCodeOK                   = ProprietaryAppCode("SUCCESS_OK")
	SuccessCodeCreated              = ProprietaryAppCode("SUCCESS_CREATED")
	SuccessCodeAccepted             = ProprietaryAppCode("SUCCESS_ACCEPTED")
	SuccessCodeNonAuthoritativeInfo = ProprietaryAppCode("SUCCESS_NON_AUTHORITATIVE_INFO")
	SuccessCodeNoContent            = ProprietaryAppCode("SUCCESS_NO_CONTENT")
	SuccessCodeResetContent         = ProprietaryAppCode("SUCCESS_RESET_CONTENT")
	SuccessCodePartialContent       = ProprietaryAppCode("SUCCESS_PARTIAL_CONTENT")
	SuccessCodeMultiStatus          = ProprietaryAppCode("SUCCESS_MULTI_STATUS")
	SuccessCodeAlreadyReported      = ProprietaryAppCode("SUCCESS_ALREADY_REPORTED")
	SuccessCodeIMUsed               = ProprietaryAppCode("SUCCESS_IM_USED")

	// 3xx Redirection
	RedirectionCodeMultipleChoices   = ProprietaryAppCode("REDIRECTION_MULTIPLE_CHOICES")
	RedirectionCodeMovedPermanently  = ProprietaryAppCode("REDIRECTION_MOVED_PERMANENTLY")
	RedirectionCodeFound             = ProprietaryAppCode("REDIRECTION_FOUND")
	RedirectionCodeSeeOther          = ProprietaryAppCode("REDIRECTION_SEE_OTHER")
	RedirectionCodeNotModified       = ProprietaryAppCode("REDIRECTION_NOT_MODIFIED")
	RedirectionCodeUseProxy          = ProprietaryAppCode("REDIRECTION_USE_PROXY")
	RedirectionCodeTemporaryRedirect = ProprietaryAppCode("REDIRECTION_TEMPORARY_REDIRECT")
	RedirectionCodePermanentRedirect = ProprietaryAppCode("REDIRECTION_PERMANENT_REDIRECT")

	// 4xx Client Errors
	ClientErrorCodeBadRequest                  = ProprietaryAppCode("CLIENT_ERROR_BAD_REQUEST")
	ClientErrorCodeUnauthorized                = ProprietaryAppCode("CLIENT_ERROR_UNAUTHORIZED")
	ClientErrorCodePaymentRequired             = ProprietaryAppCode("CLIENT_ERROR_PAYMENT_REQUIRED")
	ClientErrorCodeForbidden                   = ProprietaryAppCode("CLIENT_ERROR_FORBIDDEN")
	ClientErrorCodeNotFound                    = ProprietaryAppCode("CLIENT_ERROR_NOT_FOUND")
	ClientErrorCodeMethodNotAllowed            = ProprietaryAppCode("CLIENT_ERROR_METHOD_NOT_ALLOWED")
	ClientErrorCodeNotAcceptable               = ProprietaryAppCode("CLIENT_ERROR_NOT_ACCEPTABLE")
	ClientErrorCodeProxyAuthRequired           = ProprietaryAppCode("CLIENT_ERROR_PROXY_AUTH_REQUIRED")
	ClientErrorCodeRequestTimeout              = ProprietaryAppCode("CLIENT_ERROR_REQUEST_TIMEOUT")
	ClientErrorCodeConflict                    = ProprietaryAppCode("CLIENT_ERROR_CONFLICT")
	ClientErrorCodeGone                        = ProprietaryAppCode("CLIENT_ERROR_GONE")
	ClientErrorCodeLengthRequired              = ProprietaryAppCode("CLIENT_ERROR_LENGTH_REQUIRED")
	ClientErrorCodePreconditionFailed          = ProprietaryAppCode("CLIENT_ERROR_PRECONDITION_FAILED")
	ClientErrorCodePayloadTooLarge             = ProprietaryAppCode("CLIENT_ERROR_PAYLOAD_TOO_LARGE")
	ClientErrorCodeURITooLong                  = ProprietaryAppCode("CLIENT_ERROR_URI_TOO_LONG")
	ClientErrorCodeUnsupportedMediaType        = ProprietaryAppCode("CLIENT_ERROR_UNSUPPORTED_MEDIA_TYPE")
	ClientErrorCodeRangeNotSatisfiable         = ProprietaryAppCode("CLIENT_ERROR_RANGE_NOT_SATISFIABLE")
	ClientErrorCodeExpectationFailed           = ProprietaryAppCode("CLIENT_ERROR_EXPECTATION_FAILED")
	ClientErrorCodeTeapot                      = ProprietaryAppCode("CLIENT_ERROR_TEAPOT")
	ClientErrorCodeMisdirectedRequest          = ProprietaryAppCode("CLIENT_ERROR_MISDIRECTED_REQUEST")
	ClientErrorCodeUnprocessableEntity         = ProprietaryAppCode("CLIENT_ERROR_UNPROCESSABLE_ENTITY")
	ClientErrorCodeLocked                      = ProprietaryAppCode("CLIENT_ERROR_LOCKED")
	ClientErrorCodeFailedDependency            = ProprietaryAppCode("CLIENT_ERROR_FAILED_DEPENDENCY")
	ClientErrorCodeTooEarly                    = ProprietaryAppCode("CLIENT_ERROR_TOO_EARLY")
	ClientErrorCodeUpgradeRequired             = ProprietaryAppCode("CLIENT_ERROR_UPGRADE_REQUIRED")
	ClientErrorCodePreconditionRequired        = ProprietaryAppCode("CLIENT_ERROR_PRECONDITION_REQUIRED")
	ClientErrorCodeTooManyRequests             = ProprietaryAppCode("CLIENT_ERROR_TOO_MANY_REQUESTS")
	ClientErrorCodeRequestHeaderFieldsTooLarge = ProprietaryAppCode("CLIENT_ERROR_REQUEST_HEADER_FIELDS_TOO_LARGE")
	ClientErrorCodeUnavailableForLegalReasons  = ProprietaryAppCode("CLIENT_ERROR_UNAVAILABLE_FOR_LEGAL_REASONS")

	// 5xx Server Errors
	ServerErrorCodeInternalServerServerError     = ProprietaryAppCode("SERVER_ERROR_INTERNAL_SERVER_ERROR")
	ServerErrorCodeNotImplemented                = ProprietaryAppCode("SERVER_ERROR_NOT_IMPLEMENTED")
	ServerErrorCodeBadGateway                    = ProprietaryAppCode("SERVER_ERROR_BAD_GATEWAY")
	ServerErrorCodeServiceUnavailable            = ProprietaryAppCode("SERVER_ERROR_SERVICE_UNAVAILABLE")
	ServerErrorCodeGatewayTimeout                = ProprietaryAppCode("SERVER_ERROR_GATEWAY_TIMEOUT")
	ServerErrorCodeHTTPVersionNotSupported       = ProprietaryAppCode("SERVER_ERROR_HTTP_VERSION_NOT_SUPPORTED")
	ServerErrorCodeVariantAlsoNegotiates         = ProprietaryAppCode("SERVER_ERROR_VARIANT_ALSO_NEGOTIATES")
	ServerErrorCodeInsufficientStorage           = ProprietaryAppCode("SERVER_ERROR_INSUFFICIENT_STORAGE")
	ServerErrorCodeLoopDetected                  = ProprietaryAppCode("SERVER_ERROR_LOOP_DETECTED")
	ServerErrorCodeNotExtended                   = ProprietaryAppCode("SERVER_ERROR_NOT_EXTENDED")
	ServerErrorCodeNetworkAuthenticationRequired = ProprietaryAppCode("SERVER_ERROR_NETWORK_AUTHENTICATION_REQUIRED")
)
