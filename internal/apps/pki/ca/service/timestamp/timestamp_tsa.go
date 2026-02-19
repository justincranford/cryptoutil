// Copyright (c) 2025 Justin Cranford

// Package timestamp implements RFC 3161 Time-Stamp Protocol services.
package timestamp

import (
	crand "crypto/rand"
	"encoding/asn1"
	"fmt"
	"math/big"
	"time"
)

// PKIStatus represents the status of a timestamp response.
func (s *TSAService) CreateTimestamp(req *TimestampRequest) (*TimestampResponse, error) {
	if req == nil {
		return s.rejectResponse(PKIFailureBadRequest, "request is nil"), nil
	}

	// Validate the request.
	if err := s.validateRequest(req); err != nil {
		return s.rejectResponse(PKIFailureBadRequest, err.Error()), nil
	}

	// Generate serial number.
	serialNumber := s.generateSerialNumber()

	// Get current time.
	genTime := time.Now().UTC()

	// Build TSTInfo.
	tstInfo := TSTInfo{
		Version:        1,
		Policy:         s.getPolicy(req),
		MessageImprint: req.MessageImprint,
		SerialNumber:   serialNumber,
		GenTime:        genTime,
		Accuracy:       s.config.Accuracy,
		Ordering:       s.config.Ordering,
		Nonce:          req.Nonce,
	}

	// Build timestamp token.
	token, err := s.buildToken(&tstInfo)
	if err != nil {
		return s.rejectResponse(PKIFailureSystemFailure, err.Error()), nil
	}

	return &TimestampResponse{
		Status: PKIStatusInfo{
			Status: PKIStatusGranted,
		},
		TimeStampToken: token,
	}, nil
}

// validateRequest validates a timestamp request.
func (s *TSAService) validateRequest(req *TimestampRequest) error {
	// Validate message imprint.
	if len(req.MessageImprint.HashedMessage) == 0 {
		return fmt.Errorf("message imprint is empty")
	}

	// Validate hash algorithm.
	if !s.isAcceptedAlgorithm(req.MessageImprint.HashAlgorithm) {
		return fmt.Errorf("unsupported hash algorithm: %s", req.MessageImprint.HashAlgorithm)
	}

	// Validate hash length matches algorithm.
	expectedLen := req.MessageImprint.HashAlgorithm.CryptoHash().Size()
	if len(req.MessageImprint.HashedMessage) != expectedLen {
		return fmt.Errorf("hash length mismatch: expected %d, got %d",
			expectedLen, len(req.MessageImprint.HashedMessage))
	}

	// Validate policy if specified.
	if len(req.ReqPolicy) > 0 && !s.isAcceptedPolicy(req.ReqPolicy) {
		return fmt.Errorf("unaccepted policy: %v", req.ReqPolicy)
	}

	return nil
}

// isAcceptedAlgorithm checks if the algorithm is accepted.
func (s *TSAService) isAcceptedAlgorithm(alg HashAlgorithm) bool {
	for _, accepted := range s.config.AcceptedAlgorithms {
		if accepted == alg {
			return true
		}
	}

	return false
}

// isAcceptedPolicy checks if the policy is accepted.
func (s *TSAService) isAcceptedPolicy(policy asn1.ObjectIdentifier) bool {
	if len(s.config.AcceptedPolicies) == 0 {
		// Accept any policy if none specified.
		return true
	}

	for _, accepted := range s.config.AcceptedPolicies {
		if accepted.Equal(policy) {
			return true
		}
	}

	return false
}

// getPolicy returns the policy to use for the timestamp.
func (s *TSAService) getPolicy(req *TimestampRequest) asn1.ObjectIdentifier {
	if len(req.ReqPolicy) > 0 {
		return req.ReqPolicy
	}

	return s.config.Policy
}

// generateSerialNumber generates a unique serial number.
func (s *TSAService) generateSerialNumber() *big.Int {
	// Combine counter with random component for uniqueness.
	counter := s.serialCounter.Add(1)

	randomBytes := make([]byte, serialRandomBytes)
	_, _ = crand.Read(randomBytes)

	serial := new(big.Int).SetUint64(counter)
	serial.Lsh(serial, serialRandomBits)
	serial.Or(serial, new(big.Int).SetBytes(randomBytes))

	return serial
}

// buildToken builds the timestamp token.
func (s *TSAService) buildToken(tstInfo *TSTInfo) (*TimeStampToken, error) {
	// For now, return a simplified token.
	// Full CMS/PKCS#7 signing would be implemented for production.
	return &TimeStampToken{
		TSTInfo:    *tstInfo,
		SignedData: nil, // TODO: Implement CMS signing.
	}, nil
}

// rejectResponse creates a rejection response.
func (s *TSAService) rejectResponse(failInfo PKIFailureInfo, message string) *TimestampResponse {
	return &TimestampResponse{
		Status: PKIStatusInfo{
			Status:       PKIStatusRejection,
			StatusString: []string{message},
			FailInfo:     &failInfo,
		},
	}
}

// TimestampEntry represents a timestamp record for audit.
type TimestampEntry struct {
	SerialNumber   string    `json:"serial_number"`
	GenTime        time.Time `json:"gen_time"`
	Policy         string    `json:"policy"`
	HashAlgorithm  string    `json:"hash_algorithm"`
	HashedMessage  string    `json:"hashed_message"`
	Nonce          string    `json:"nonce,omitempty"`
	TSACertificate string    `json:"tsa_certificate"`
}

// ToEntry converts a TSTInfo to a TimestampEntry for auditing.
func (t *TSTInfo) ToEntry(tsaCert string) *TimestampEntry {
	entry := &TimestampEntry{
		SerialNumber:   t.SerialNumber.Text(hexBase),
		GenTime:        t.GenTime,
		Policy:         t.Policy.String(),
		HashAlgorithm:  string(t.MessageImprint.HashAlgorithm),
		HashedMessage:  fmt.Sprintf("%x", t.MessageImprint.HashedMessage),
		TSACertificate: tsaCert,
	}

	if t.Nonce != nil {
		entry.Nonce = t.Nonce.Text(hexBase)
	}

	return entry
}

// Constants.
const (
	hexBase           = 16
	serialRandomBytes = 8
	serialRandomBits  = 64
	unknownStatus     = "unknown"
)

// ASN.1 structures for RFC 3161 TimeStampReq parsing.

// timeStampReqASN1 is the ASN.1 structure for TimeStampReq.
type timeStampReqASN1 struct {
	Version        int
	MessageImprint messageImprintASN1
	ReqPolicy      asn1.ObjectIdentifier `asn1:"optional"`
	Nonce          *big.Int              `asn1:"optional"`
	CertReq        bool                  `asn1:"optional,default:false"`
	Extensions     []extensionASN1       `asn1:"optional,tag:0"`
}

// messageImprintASN1 is the ASN.1 structure for MessageImprint.
type messageImprintASN1 struct {
	HashAlgorithm algorithmIdentifierASN1
	HashedMessage []byte
}

// algorithmIdentifierASN1 is the ASN.1 structure for AlgorithmIdentifier.
type algorithmIdentifierASN1 struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

// extensionASN1 is the ASN.1 structure for X.509 Extension.
type extensionASN1 struct {
	OID      asn1.ObjectIdentifier
	Critical bool `asn1:"optional,default:false"`
	Value    []byte
}

// ParseTimestampRequest parses a DER-encoded RFC 3161 TimeStampReq.
func ParseTimestampRequest(der []byte) (*TimestampRequest, error) {
	if len(der) == 0 {
		return nil, fmt.Errorf("empty timestamp request")
	}

	var req timeStampReqASN1

	rest, err := asn1.Unmarshal(der, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp request: %w", err)
	}

	if len(rest) > 0 {
		return nil, fmt.Errorf("trailing data after timestamp request")
	}

	// Convert ASN.1 hash algorithm OID to our HashAlgorithm type.
	hashAlg, err := oidToHashAlgorithm(req.MessageImprint.HashAlgorithm.Algorithm)
	if err != nil {
		return nil, err
	}

	// Convert extensions.
	extensions := make([]Extension, 0, len(req.Extensions))

	for _, ext := range req.Extensions {
		extensions = append(extensions, Extension(ext))
	}

	return &TimestampRequest{
		Version: req.Version,
		MessageImprint: MessageImprint{
			HashAlgorithm: hashAlg,
			HashedMessage: req.MessageImprint.HashedMessage,
		},
		ReqPolicy:  req.ReqPolicy,
		Nonce:      req.Nonce,
		CertReq:    req.CertReq,
		Extensions: extensions,
	}, nil
}

// oidToHashAlgorithm converts an ASN.1 OID to a HashAlgorithm.
func oidToHashAlgorithm(oid asn1.ObjectIdentifier) (HashAlgorithm, error) {
	oidAlgMap := map[string]HashAlgorithm{
		"2.16.840.1.101.3.4.2.1": HashAlgorithmSHA256,
		"2.16.840.1.101.3.4.2.2": HashAlgorithmSHA384,
		"2.16.840.1.101.3.4.2.3": HashAlgorithmSHA512,
	}

	if alg, ok := oidAlgMap[oid.String()]; ok {
		return alg, nil
	}

	return "", fmt.Errorf("unsupported hash algorithm OID: %v", oid)
}

// timeStampRespASN1 is the ASN.1 structure for TimeStampResp.
type timeStampRespASN1 struct {
	Status         pkiStatusInfoASN1
	TimeStampToken asn1.RawValue `asn1:"optional"`
}

// pkiStatusInfoASN1 is the ASN.1 structure for PKIStatusInfo.
type pkiStatusInfoASN1 struct {
	Status       int
	StatusString []string       `asn1:"optional,utf8"`
	FailInfo     asn1.BitString `asn1:"optional"`
}

// tstInfoASN1 is the ASN.1 structure for TSTInfo.
type tstInfoASN1 struct {
	Version        int
	Policy         asn1.ObjectIdentifier
	MessageImprint messageImprintASN1
	SerialNumber   *big.Int
	GenTime        time.Time `asn1:"generalized"`
	Accuracy       accuracyASN1
	Ordering       bool            `asn1:"optional,default:false"`
	Nonce          *big.Int        `asn1:"optional"`
	TSA            asn1.RawValue   `asn1:"optional,tag:0"`
	Extensions     []extensionASN1 `asn1:"optional,tag:1"`
}

// accuracyASN1 is the ASN.1 structure for Accuracy.
type accuracyASN1 struct {
	Seconds int `asn1:"optional"`
	Millis  int `asn1:"optional,tag:0"`
	Micros  int `asn1:"optional,tag:1"`
}

// SerializeTimestampResponse serializes a TimestampResponse to DER format.
func SerializeTimestampResponse(resp *TimestampResponse) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	// Build PKIStatusInfo.
	statusInfo := pkiStatusInfoASN1{
		Status:       int(resp.Status.Status),
		StatusString: resp.Status.StatusString,
	}

	// Add failure info if present.
	if resp.Status.FailInfo != nil {
		// PKI failure info is a bit string with the failure bit set.
		failBit := int(*resp.Status.FailInfo)
		// Create a bit string with the appropriate bit set.
		failBytes := make([]byte, (failBit/bitStringBitsPerByte)+1)

		if len(failBytes) > 0 {
			failBytes[failBit/bitStringBitsPerByte] = 1 << (bitStringBitShift - (failBit % bitStringBitsPerByte))
		}

		statusInfo.FailInfo = asn1.BitString{Bytes: failBytes, BitLength: failBit + 1}
	}

	// Build response.
	response := timeStampRespASN1{
		Status: statusInfo,
	}

	// Add timestamp token if present.
	if resp.TimeStampToken != nil {
		// Build TSTInfo.
		tstInfo := tstInfoASN1{
			Version: resp.TimeStampToken.TSTInfo.Version,
			Policy:  resp.TimeStampToken.TSTInfo.Policy,
			MessageImprint: messageImprintASN1{
				HashAlgorithm: algorithmIdentifierASN1{
					Algorithm: resp.TimeStampToken.TSTInfo.MessageImprint.HashAlgorithm.OID(),
				},
				HashedMessage: resp.TimeStampToken.TSTInfo.MessageImprint.HashedMessage,
			},
			SerialNumber: resp.TimeStampToken.TSTInfo.SerialNumber,
			GenTime:      resp.TimeStampToken.TSTInfo.GenTime,
			Ordering:     resp.TimeStampToken.TSTInfo.Ordering,
			Nonce:        resp.TimeStampToken.TSTInfo.Nonce,
		}

		// Add accuracy if present.
		if resp.TimeStampToken.TSTInfo.Accuracy != nil {
			tstInfo.Accuracy = accuracyASN1{
				Seconds: resp.TimeStampToken.TSTInfo.Accuracy.Seconds,
				Millis:  resp.TimeStampToken.TSTInfo.Accuracy.Millis,
				Micros:  resp.TimeStampToken.TSTInfo.Accuracy.Micros,
			}
		}

		// Serialize TSTInfo to DER.
		tstInfoDER, err := asn1.Marshal(tstInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal TSTInfo: %w", err)
		}

		// For a complete implementation, this would be wrapped in CMS SignedData.
		// For now, we embed the TSTInfo directly for testing purposes.
		response.TimeStampToken = asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      tstInfoDER,
		}
	}

	der, err := asn1.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal timestamp response: %w", err)
	}

	return der, nil
}

// Bit string constants.
const (
	bitStringBitsPerByte = 8
	bitStringBitShift    = 7
)
