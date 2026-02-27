// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"time"
)

// AuthContext represents the contextual information for an authentication attempt.
type AuthContext struct {
	Location *GeoLocation       // Geographic location.
	Device   *DeviceFingerprint // Device information.
	Time     time.Time          // Time of authentication attempt.
	Network  *NetworkInfo       // Network information.
	Behavior *UserBehavior      // User behavior patterns.
}

// GeoLocation represents geographic location information.
type GeoLocation struct {
	Country   string
	City      string
	Region    string
	Latitude  float64
	Longitude float64
	IPAddress string
}

// DeviceFingerprint represents device identification information.
type DeviceFingerprint struct {
	ID           string            // Unique device identifier.
	UserAgent    string            // Browser user agent.
	OS           string            // Operating system.
	Browser      string            // Browser name and version.
	ScreenSize   string            // Screen resolution.
	Timezone     string            // Device timezone.
	Language     string            // Browser language.
	Plugins      []string          // Installed plugins.
	Fonts        []string          // Installed fonts.
	Canvas       string            // Canvas fingerprint.
	WebGL        string            // WebGL fingerprint.
	AudioContext string            // Audio context fingerprint.
	Metadata     map[string]string // Additional metadata.
}

// NetworkInfo represents network connection information.
type NetworkInfo struct {
	IPAddress string
	ISP       string
	ASN       string
	IsVPN     bool
	IsProxy   bool
	IsTor     bool
}

// UserBehavior represents behavioral patterns for analysis.
type UserBehavior struct {
	TypingSpeed     float64         // Average typing speed (chars/sec).
	MouseMovements  []MouseMovement // Mouse movement patterns.
	ClickPatterns   []ClickPattern  // Click behavior patterns.
	NavigationFlow  []string        // Page navigation sequence.
	SessionDuration time.Duration   // Session duration.
	Metadata        map[string]any  // Additional behavioral data.
}

// MouseMovement represents a mouse movement data point.
type MouseMovement struct {
	X         int
	Y         int
	Timestamp time.Time
}

// ClickPattern represents a click behavior data point.
type ClickPattern struct {
	X         int
	Y         int
	Timestamp time.Time
	Button    string // "left", "right", "middle".
}

// UserBaseline represents the baseline behavior profile for a user.
type UserBaseline struct {
	UserID          string
	KnownLocations  []GeoLocation
	KnownDevices    []string
	KnownNetworks   []string
	TypicalHours    []int
	BehaviorProfile *BehaviorProfile
	LastAuthTime    time.Time
	TotalAuthCount  int
	FailedAuthCount int
}

// BehaviorProfile represents learned behavioral patterns.
type BehaviorProfile struct {
	AverageTypingSpeed   float64
	AverageSessionLength time.Duration
	CommonNavigationFlow []string
	PreferredLanguage    string
	PreferredTimezone    string
}

// Anomaly represents a detected behavioral anomaly.
type Anomaly struct {
	Type        string
	Severity    float64
	Description string
	Metadata    map[string]any
}

// ContextAnalyzer analyzes authentication context for anomalies.
type ContextAnalyzer interface {
	AnalyzeContext(ctx context.Context, request *AuthRequest) (*AuthContext, error)
	DetectAnomalies(ctx context.Context, authContext *AuthContext, baseline *UserBaseline) ([]Anomaly, error)
}

// AuthRequest represents an authentication request with contextual data.
type AuthRequest struct {
	UserID    string
	IPAddress string
	UserAgent string
	Timestamp time.Time
	Headers   map[string]string
	Metadata  map[string]any
}

// DefaultContextAnalyzer implements context analysis.
type DefaultContextAnalyzer struct {
	geoIP    GeoIPService
	deviceDB DeviceFingerprintDB
}

// NewDefaultContextAnalyzer creates a new context analyzer.
func NewDefaultContextAnalyzer(geoIP GeoIPService, deviceDB DeviceFingerprintDB) *DefaultContextAnalyzer {
	return &DefaultContextAnalyzer{
		geoIP:    geoIP,
		deviceDB: deviceDB,
	}
}

// AnalyzeContext extracts authentication context from the request.
func (a *DefaultContextAnalyzer) AnalyzeContext(ctx context.Context, request *AuthRequest) (*AuthContext, error) {
	authContext := &AuthContext{
		Time: request.Timestamp,
	}

	// Get geo location from IP.
	if a.geoIP != nil {
		location, err := a.geoIP.Lookup(ctx, request.IPAddress)
		if err == nil {
			authContext.Location = location
		}
	}

	// Get device fingerprint from user agent and headers.
	if a.deviceDB != nil {
		device, err := a.deviceDB.GetFingerprint(ctx, request.UserAgent, request.Headers)
		if err == nil {
			authContext.Device = device
		}
	}

	// Extract network info.
	authContext.Network = &NetworkInfo{
		IPAddress: request.IPAddress,
		// Additional network analysis would go here.
	}

	return authContext, nil
}

// DetectAnomalies identifies anomalous patterns in authentication context.
func (a *DefaultContextAnalyzer) DetectAnomalies(_ context.Context, authContext *AuthContext, baseline *UserBaseline) ([]Anomaly, error) {
	anomalies := make([]Anomaly, 0)

	// Check for location anomalies.
	if authContext.Location != nil {
		locationAnomaly := a.detectLocationAnomaly(authContext.Location, baseline)
		if locationAnomaly != nil {
			anomalies = append(anomalies, *locationAnomaly)
		}
	}

	// Check for device anomalies.
	if authContext.Device != nil {
		deviceAnomaly := a.detectDeviceAnomaly(authContext.Device, baseline)
		if deviceAnomaly != nil {
			anomalies = append(anomalies, *deviceAnomaly)
		}
	}

	// Check for time-based anomalies.
	timeAnomaly := a.detectTimeAnomaly(authContext.Time, baseline)
	if timeAnomaly != nil {
		anomalies = append(anomalies, *timeAnomaly)
	}

	// Check for velocity anomalies.
	if !baseline.LastAuthTime.IsZero() {
		velocityAnomaly := a.detectVelocityAnomaly(authContext.Time, baseline.LastAuthTime)
		if velocityAnomaly != nil {
			anomalies = append(anomalies, *velocityAnomaly)
		}
	}

	return anomalies, nil
}

// Helper methods for anomaly detection.

func (a *DefaultContextAnalyzer) detectLocationAnomaly(location *GeoLocation, baseline *UserBaseline) *Anomaly {
	for _, knownLoc := range baseline.KnownLocations {
		if knownLoc.Country == location.Country && knownLoc.City == location.City {
			return nil // Known location, no anomaly.
		}
	}

	const severityUnknownLocation = 0.7

	return &Anomaly{
		Type:        "unknown_location",
		Severity:    severityUnknownLocation,
		Description: "Authentication from unknown location",
		Metadata: map[string]any{
			cryptoutilSharedMagic.AddressCountry: location.Country,
			"city":                               location.City,
		},
	}
}

func (a *DefaultContextAnalyzer) detectDeviceAnomaly(device *DeviceFingerprint, baseline *UserBaseline) *Anomaly {
	for _, knownDevice := range baseline.KnownDevices {
		if knownDevice == device.ID {
			return nil // Known device, no anomaly.
		}
	}

	const severityUnknownDevice = 0.6

	return &Anomaly{
		Type:        "unknown_device",
		Severity:    severityUnknownDevice,
		Description: "Authentication from unknown device",
		Metadata: map[string]any{
			"device_id": device.ID,
		},
	}
}

func (a *DefaultContextAnalyzer) detectTimeAnomaly(authTime time.Time, baseline *UserBaseline) *Anomaly {
	hour := authTime.Hour()

	for _, typicalHour := range baseline.TypicalHours {
		if hour == typicalHour {
			return nil // Typical time, no anomaly.
		}
	}

	const severityUnusualTime = 0.4

	return &Anomaly{
		Type:        "unusual_time",
		Severity:    severityUnusualTime,
		Description: "Authentication at unusual time",
		Metadata: map[string]any{
			"hour": hour,
		},
	}
}

func (a *DefaultContextAnalyzer) detectVelocityAnomaly(authTime, lastAuthTime time.Time) *Anomaly {
	const (
		veryFastThreshold = 5 * time.Second
		severityVeryFast  = 0.9
	)

	timeSinceLastAuth := authTime.Sub(lastAuthTime)

	if timeSinceLastAuth < veryFastThreshold {
		return &Anomaly{
			Type:        "high_velocity",
			Severity:    severityVeryFast,
			Description: "Extremely fast authentication attempts",
			Metadata: map[string]any{
				"time_since_last_auth": timeSinceLastAuth.String(),
			},
		}
	}

	return nil
}

// Supporting service interfaces.

// GeoIPService provides geographic lookup for IP addresses.
type GeoIPService interface {
	Lookup(ctx context.Context, ipAddress string) (*GeoLocation, error)
}

// DeviceFingerprintDB provides device fingerprinting capabilities.
type DeviceFingerprintDB interface {
	GetFingerprint(ctx context.Context, userAgent string, headers map[string]string) (*DeviceFingerprint, error)
	StoreFingerprint(ctx context.Context, fingerprint *DeviceFingerprint) error
}

// UserBehaviorStore stores and retrieves user behavioral data.
type UserBehaviorStore interface {
	GetBaseline(ctx context.Context, userID string) (*UserBaseline, error)
	UpdateBaseline(ctx context.Context, userID string, authContext *AuthContext) error
	RecordAuthentication(ctx context.Context, userID string, success bool, authContext *AuthContext) error
}
