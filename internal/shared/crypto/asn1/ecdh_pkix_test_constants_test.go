// Copyright (c) 2025-2026 Justin Cranford.
//

package asn1

const (
	testECDHASN1MarshalErrorCase            = "ASN1 marshal error"
	testECDHInjectedASN1MarshalError        = "injected ASN1 marshal error"
	testECDHInjectedMarshalError            = "injected marshal error"
	testECDHInjectedOIDParseError           = "injected OID parse error"
	testECDHInvalidECDHOIDCase              = "invalid ECDH OID"
	testECDHInvalidECAlgorithmOIDCase       = "invalid EC OID"
	testECDHInvalidObjectIdentifierArcError = "invalid object identifier arc"
	testECDHMarshalErrorCase                = "marshal error"
	testECDHMissingCurveParametersCase      = "missing curve parameters"
	testECDHOrdinaryECDSACompatibilityCase  = "ordinary ECDSA remains supported"
	testECDHParseECAlgorithmOIDError        = "parse id-ecPublicKey OID"
	testECDHParseECDHAlgorithmOIDError      = "parse id-ecDH OID"
	testECDHTrailingDataCase                = "trailing ECDH data"
	testECDHUnexpectedAlgorithmOIDCase      = "unexpected algorithm OID"
	testECDHUnexpectedParsedKeyTypeCase     = "unexpected parsed key type"
	testECDHUnsupportedCurveCase            = "unsupported ECDH curve"
	testECDHInvalidASN1SyntaxError          = "syntax error"
	testECDHOIDZero                         = "0.0"
	testECDHOIDOne                          = "1.39"
	testECDHOIDTwo                          = "2.999"
)
