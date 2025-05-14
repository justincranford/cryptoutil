package jose

import (
	"fmt"

	joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

func LogJweInfo(jwsMessage *joseJws.Message) error {
	if jwsMessage == nil {
		return fmt.Errorf("jwsMessage is nil")
	} else if len(jwsMessage.Signatures()) == 0 {
		return fmt.Errorf("jwsMessage has no signatures")
	}

	for i, jwsSignature := range jwsMessage.Signatures() {
		logMessageSigHeader := fmt.Sprintf("JWS Header[%d]:", i)

		protectedHeaders := jwsSignature.ProtectedHeaders()
		if alg, ok := protectedHeaders.Algorithm(); ok {
			logMessageSigHeader += fmt.Sprintf(" alg=%s\n", alg)
		}
		if kid, ok := protectedHeaders.KeyID(); ok {
			logMessageSigHeader += fmt.Sprintf(" kid=%s\n", kid)
		}
		if typ, ok := protectedHeaders.Type(); ok {
			logMessageSigHeader += fmt.Sprintf(" typ=%s\n", typ)
		}
		if cty, ok := protectedHeaders.ContentType(); ok {
			logMessageSigHeader += fmt.Sprintf(" cty=%s\n", cty)
		}
		if jku, ok := protectedHeaders.JWKSetURL(); ok {
			logMessageSigHeader += fmt.Sprintf(" jku=%s\n", jku)
		}
		if x5u, ok := protectedHeaders.X509URL(); ok {
			logMessageSigHeader += fmt.Sprintf(" x5u=%s\n", x5u)
		}
		if x5c, ok := protectedHeaders.X509CertChain(); ok {
			logMessageSigHeader += fmt.Sprintf(" x5c=%v\n", x5c)
		}
		if x5t, ok := protectedHeaders.X509CertThumbprint(); ok {
			logMessageSigHeader += fmt.Sprintf(" x5t=%s\n", x5t)
		}
		if x5tS256, ok := protectedHeaders.X509CertThumbprintS256(); ok {
			logMessageSigHeader += fmt.Sprintf(" x5t#S256=%s\n", x5tS256)
		}
		if crit, ok := protectedHeaders.Critical(); ok {
			logMessageSigHeader += fmt.Sprintf(" crit=%v\n", crit)
		}

		publicHeaders := jwsSignature.PublicHeaders()
		for _, key := range publicHeaders.Keys() {
			var value any
			err := publicHeaders.Get(key, &value)
			if err != nil {
				logMessageSigHeader += fmt.Sprintf(" %s=%v\n", key, value)
			}
		}

		fmt.Print(logMessageSigHeader)
	}

	fmt.Printf("JWS Payload: %s\n", string(jwsMessage.Payload()))

	for i, jwsSignature := range jwsMessage.Signatures() {
		fmt.Printf("JWS Signature[%d]: %s\n", i, jwsSignature.Signature())
	}

	return nil
}
