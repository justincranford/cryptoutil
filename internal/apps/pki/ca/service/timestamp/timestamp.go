// Copyright (c) 2025 Justin Cranford

// Package timestamp implements RFC 3161 Time-Stamp Protocol services.
package timestamp

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
)

// PKIStatus represents the status of a timestamp response.
type PKIStatus int

// PKI status codes per RFC 3161.
const (
	PKIStatusGranted                PKIStatus = 0
	PKIStatusGrantedWithMods        PKIStatus = 1
	PKIStatusRejection              PKIStatus = 2
	PKIStatusWaiting                PKIStatus = 3
	PKIStatusRevocationWarning      PKIStatus = 4
	PKIStatusRevocationNotification PKIStatus = 5
)

// String returns the string representation of a PKI status.
func (s PKIStatus) String() string {
	statusStrings := map[PKIStatus]string{
		PKIStatusGranted:                "granted",
		PKIStatusGrantedWithMods:        "grantedWithMods",
		PKIStatusRejection:              "rejection",
		PKIStatusWaiting:                "waiting",
		PKIStatusRevocationWarning:      "revocationWarning",
		PKIStatusRevocationNotification: "revocationNotification",
	}

	if str, ok := statusStrings[s]; ok {
		return str
	}

	return unknownStatus
}

// PKIFailureInfo represents failure information in a timestamp response.
type PKIFailureInfo int

// PKI failure codes per RFC 3161.
const (
	PKIFailureBadAlg              PKIFailureInfo = 0
	PKIFailureBadRequest          PKIFailureInfo = 2
	PKIFailureBadDataFormat       PKIFailureInfo = 5
	PKIFailureTimeNotAvailable    PKIFailureInfo = 14
	PKIFailureUnacceptedPolicy    PKIFailureInfo = 15
	PKIFailureUnacceptedExtension PKIFailureInfo = 16
	PKIFailureAddInfoNotAvailable PKIFailureInfo = 17
	PKIFailureSystemFailure       PKIFailureInfo = 25
)

// String returns the string representation of a PKI failure info.
func (f PKIFailureInfo) String() string {
	failureStrings := map[PKIFailureInfo]string{
		PKIFailureBadAlg:              "badAlg",
		PKIFailureBadRequest:          "badRequest",
		PKIFailureBadDataFormat:       "badDataFormat",
		PKIFailureTimeNotAvailable:    "timeNotAvailable",
		PKIFailureUnacceptedPolicy:    "unacceptedPolicy",
		PKIFailureUnacceptedExtension: "unacceptedExtension",
		PKIFailureAddInfoNotAvailable: "addInfoNotAvailable",
		PKIFailureSystemFailure:       "systemFailure",
	}

	if str, ok := failureStrings[f]; ok {
		return str
	}

	return unknownStatus
}

// HashAlgorithm represents supported hash algorithms for timestamping.
type HashAlgorithm string

// Supported hash algorithms.
const (
	HashAlgorithmSHA256 HashAlgorithm = "SHA-256"
	HashAlgorithmSHA384 HashAlgorithm = "SHA-384"
	HashAlgorithmSHA512 HashAlgorithm = "SHA-512"
)

// OID returns the ASN.1 OID for the hash algorithm.
func (h HashAlgorithm) OID() asn1.ObjectIdentifier {
	hashOIDs := map[HashAlgorithm]asn1.ObjectIdentifier{
		HashAlgorithmSHA256: {2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 1},
		HashAlgorithmSHA384: {2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 2},
		HashAlgorithmSHA512: {2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 3},
	}

	if oid, ok := hashOIDs[h]; ok {
		return oid
	}

	return nil
}

// CryptoHash returns the crypto.Hash for the algorithm.
func (h HashAlgorithm) CryptoHash() crypto.Hash {
	cryptoHashes := map[HashAlgorithm]crypto.Hash{
		HashAlgorithmSHA256: crypto.SHA256,
		HashAlgorithmSHA384: crypto.SHA384,
		HashAlgorithmSHA512: crypto.SHA512,
	}

	if hash, ok := cryptoHashes[h]; ok {
		return hash
	}

	return 0
}

// TimestampRequest represents an RFC 3161 timestamp request.
type TimestampRequest struct {
	// Version is the request version (typically 1).
	Version int

	// MessageImprint contains the hash of the data to be timestamped.
	MessageImprint MessageImprint

	// ReqPolicy is the requested TSA policy OID.
	ReqPolicy asn1.ObjectIdentifier

	// Nonce is an optional nonce for replay protection.
	Nonce *big.Int

	// CertReq indicates if the TSA certificate should be included.
	CertReq bool

	// Extensions contains any requested extensions.
	Extensions []Extension
}

// MessageImprint contains the hash algorithm and hash value.
type MessageImprint struct {
	// HashAlgorithm identifies the hash algorithm used.
	HashAlgorithm HashAlgorithm

	// HashedMessage is the hash of the data.
	HashedMessage []byte
}

// Extension represents an X.509 extension.
type Extension struct {
	OID      asn1.ObjectIdentifier
	Critical bool
	Value    []byte
}

// TimestampResponse represents an RFC 3161 timestamp response.
type TimestampResponse struct {
	// Status contains the PKI status info.
	Status PKIStatusInfo

	// TimeStampToken contains the signed timestamp token (if granted).
	TimeStampToken *TimeStampToken
}

// PKIStatusInfo contains the status of the timestamp request.
type PKIStatusInfo struct {
	// Status is the PKI status code.
	Status PKIStatus

	// StatusString contains additional status information.
	StatusString []string

	// FailInfo contains failure information if status is rejection.
	FailInfo *PKIFailureInfo
}

// TimeStampToken contains the timestamp token information.
type TimeStampToken struct {
	// TSTInfo contains the timestamp token info.
	TSTInfo TSTInfo

	// SignedData contains the CMS signed data (DER encoded).
	SignedData []byte
}

// TSTInfo contains the timestamp token information per RFC 3161.
type TSTInfo struct {
	// Version is the TSTInfo version (typically 1).
	Version int

	// Policy is the TSA policy OID.
	Policy asn1.ObjectIdentifier

	// MessageImprint is the hash that was timestamped.
	MessageImprint MessageImprint

	// SerialNumber is the unique serial number for this timestamp.
	SerialNumber *big.Int

	// GenTime is the time the timestamp was generated.
	GenTime time.Time

	// Accuracy contains optional accuracy information.
	Accuracy *Accuracy

	// Ordering indicates if timestamps from this TSA are ordered.
	Ordering bool

	// Nonce is the nonce from the request (if present).
	Nonce *big.Int

	// TSA is the name of the TSA.
	TSA *GeneralName

	// Extensions contains any extensions.
	Extensions []Extension
}

// Accuracy represents the accuracy of the timestamp.
type Accuracy struct {
	// Seconds is the accuracy in seconds.
	Seconds int

	// Millis is the additional accuracy in milliseconds.
	Millis int

	// Micros is the additional accuracy in microseconds.
	Micros int
}

// GeneralName represents an X.509 GeneralName.
type GeneralName struct {
	// Type indicates the type of name.
	Type int

	// Value is the name value.
	Value string
}

// TSAConfig configures the Time-Stamp Authority service.
type TSAConfig struct {
	// Certificate is the TSA's signing certificate.
	Certificate *x509.Certificate

	// PrivateKey is the TSA's signing key.
	PrivateKey crypto.Signer

	// Provider handles cryptographic operations.
	Provider cryptoutilCACrypto.Provider

	// Policy is the TSA's policy OID.
	Policy asn1.ObjectIdentifier

	// AcceptedPolicies lists policies this TSA accepts.
	AcceptedPolicies []asn1.ObjectIdentifier

	// AcceptedAlgorithms lists hash algorithms this TSA accepts.
	AcceptedAlgorithms []HashAlgorithm

	// Accuracy defines the timestamp accuracy.
	Accuracy *Accuracy

	// Ordering indicates if this TSA provides ordered timestamps.
	Ordering bool

	// IncludeCertificate indicates if the TSA cert should be included.
	IncludeCertificate bool
}

// TSAService implements the Time-Stamp Authority service.
type TSAService struct {
	config        *TSAConfig
	serialCounter atomic.Uint64
}

// NewTSAService creates a new Time-Stamp Authority service.
func NewTSAService(config *TSAConfig) (*TSAService, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if config.Certificate == nil {
		return nil, fmt.Errorf("certificate is required")
	}

	if config.PrivateKey == nil {
		return nil, fmt.Errorf("private key is required")
	}

	if config.Provider == nil {
		return nil, fmt.Errorf("crypto provider is required")
	}

	if len(config.Policy) == 0 {
		return nil, fmt.Errorf("policy OID is required")
	}

	if len(config.AcceptedAlgorithms) == 0 {
		// Default to SHA-256 if not specified.
		config.AcceptedAlgorithms = []HashAlgorithm{HashAlgorithmSHA256}
	}

	return &TSAService{
		config: config,
	}, nil
}

// CreateTimestamp processes a timestamp request and returns a response.
