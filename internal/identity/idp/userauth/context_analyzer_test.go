// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// contextTestGeoIPService is a mock implementation of GeoIPService for context analyzer tests.
type contextTestGeoIPService struct {
	location *GeoLocation
	err      error
}

func (m *contextTestGeoIPService) Lookup(_ context.Context, _ string) (*GeoLocation, error) {
	return m.location, m.err
}

// contextTestDeviceFingerprintDB is a mock implementation of DeviceFingerprintDB for context analyzer tests.
type contextTestDeviceFingerprintDB struct {
	device *DeviceFingerprint
	err    error
}

func (m *contextTestDeviceFingerprintDB) GetFingerprint(_ context.Context, _ string, _ map[string]string) (*DeviceFingerprint, error) {
	return m.device, m.err
}

func (m *contextTestDeviceFingerprintDB) StoreFingerprint(_ context.Context, _ *DeviceFingerprint) error {
	return m.err
}

func TestNewDefaultContextAnalyzer(t *testing.T) {
	t.Parallel()

	geoIP := &contextTestGeoIPService{}
	deviceDB := &contextTestDeviceFingerprintDB{}

	analyzer := NewDefaultContextAnalyzer(geoIP, deviceDB)
	require.NotNil(t, analyzer)
}

func TestAnalyzeContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		geoIP     *contextTestGeoIPService
		deviceDB  *contextTestDeviceFingerprintDB
		request   *AuthRequest
		expectGeo bool
		expectDev bool
		expectNet bool
	}{
		{
			name: "with_all_services",
			geoIP: &contextTestGeoIPService{
				location: &GeoLocation{
					Country: "US",
					City:    "New York",
				},
			},
			deviceDB: &contextTestDeviceFingerprintDB{
				device: &DeviceFingerprint{
					ID:        "device-123",
					UserAgent: "Mozilla/5.0",
				},
			},
			request: &AuthRequest{
				UserID:    "user-1",
				IPAddress: "192.168.1.1",
				UserAgent: "Mozilla/5.0",
				Timestamp: time.Now().UTC(),
			},
			expectGeo: true,
			expectDev: true,
			expectNet: true,
		},
		{
			name:  "with_nil_services",
			geoIP: nil,
			request: &AuthRequest{
				UserID:    "user-1",
				IPAddress: "192.168.1.1",
				Timestamp: time.Now().UTC(),
			},
			expectGeo: false,
			expectDev: false,
			expectNet: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var analyzer *DefaultContextAnalyzer
			if tc.geoIP != nil || tc.deviceDB != nil {
				analyzer = NewDefaultContextAnalyzer(tc.geoIP, tc.deviceDB)
			} else {
				analyzer = NewDefaultContextAnalyzer(nil, nil)
			}

			ctx := context.Background()
			authContext, err := analyzer.AnalyzeContext(ctx, tc.request)

			require.NoError(t, err)
			require.NotNil(t, authContext)
			require.Equal(t, tc.request.Timestamp, authContext.Time)

			if tc.expectGeo {
				require.NotNil(t, authContext.Location)
			}

			if tc.expectDev {
				require.NotNil(t, authContext.Device)
			}

			if tc.expectNet {
				require.NotNil(t, authContext.Network)
				require.Equal(t, tc.request.IPAddress, authContext.Network.IPAddress)
			}
		})
	}
}

func TestDetectAnomalies(t *testing.T) {
	t.Parallel()

	analyzer := NewDefaultContextAnalyzer(nil, nil)
	ctx := context.Background()

	tests := []struct {
		name            string
		authContext     *AuthContext
		baseline        *UserBaseline
		expectAnomalies int
		anomalyTypes    []string
	}{
		{
			name: "no_anomalies_known_location_and_device",
			authContext: &AuthContext{
				Location: &GeoLocation{
					Country: "US",
					City:    "New York",
				},
				Device: &DeviceFingerprint{
					ID: "device-123",
				},
				Time: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				KnownLocations: []GeoLocation{
					{Country: "US", City: "New York"},
				},
				KnownDevices: []string{"device-123"},
				TypicalHours: []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
			},
			expectAnomalies: 0,
		},
		{
			name: "unknown_location",
			authContext: &AuthContext{
				Location: &GeoLocation{
					Country: "RU",
					City:    "Moscow",
				},
				Time: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				KnownLocations: []GeoLocation{
					{Country: "US", City: "New York"},
				},
				TypicalHours: []int{10},
			},
			expectAnomalies: 1,
			anomalyTypes:    []string{"unknown_location"},
		},
		{
			name: "unknown_device",
			authContext: &AuthContext{
				Device: &DeviceFingerprint{
					ID: "new-device-456",
				},
				Time: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				KnownDevices: []string{"device-123"},
				TypicalHours: []int{10},
			},
			expectAnomalies: 1,
			anomalyTypes:    []string{"unknown_device"},
		},
		{
			name: "unusual_time",
			authContext: &AuthContext{
				Time: time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				TypicalHours: []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
			},
			expectAnomalies: 1,
			anomalyTypes:    []string{"unusual_time"},
		},
		{
			name: "high_velocity",
			authContext: &AuthContext{
				Time: time.Date(2025, 1, 1, 10, 0, 3, 0, time.UTC),
			},
			baseline: &UserBaseline{
				TypicalHours: []int{10},
				LastAuthTime: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			expectAnomalies: 1,
			anomalyTypes:    []string{"high_velocity"},
		},
		{
			name: "multiple_anomalies",
			authContext: &AuthContext{
				Location: &GeoLocation{
					Country: "CN",
					City:    "Beijing",
				},
				Device: &DeviceFingerprint{
					ID: "unknown-device",
				},
				Time: time.Date(2025, 1, 1, 3, 0, 3, 0, time.UTC),
			},
			baseline: &UserBaseline{
				KnownLocations: []GeoLocation{
					{Country: "US", City: "New York"},
				},
				KnownDevices: []string{"device-123"},
				TypicalHours: []int{9, 10, 11},
				LastAuthTime: time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC),
			},
			expectAnomalies: 4,
			anomalyTypes:    []string{"unknown_location", "unknown_device", "unusual_time", "high_velocity"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			anomalies, err := analyzer.DetectAnomalies(ctx, tc.authContext, tc.baseline)
			require.NoError(t, err)
			require.Len(t, anomalies, tc.expectAnomalies)

			if tc.anomalyTypes != nil {
				for i, expectedType := range tc.anomalyTypes {
					require.Equal(t, expectedType, anomalies[i].Type)
				}
			}
		})
	}
}

func TestAnomalySeverities(t *testing.T) {
	t.Parallel()

	analyzer := NewDefaultContextAnalyzer(nil, nil)
	ctx := context.Background()

	tests := []struct {
		name             string
		authContext      *AuthContext
		baseline         *UserBaseline
		expectedSeverity float64
		anomalyType      string
	}{
		{
			name: "unknown_location_severity",
			authContext: &AuthContext{
				Location: &GeoLocation{
					Country: "RU",
					City:    "Moscow",
				},
				Time: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				KnownLocations: []GeoLocation{},
				TypicalHours:   []int{10},
			},
			expectedSeverity: 0.7,
			anomalyType:      "unknown_location",
		},
		{
			name: "unknown_device_severity",
			authContext: &AuthContext{
				Device: &DeviceFingerprint{
					ID: "new-device",
				},
				Time: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				KnownDevices: []string{},
				TypicalHours: []int{10},
			},
			expectedSeverity: 0.6,
			anomalyType:      "unknown_device",
		},
		{
			name: "unusual_time_severity",
			authContext: &AuthContext{
				Time: time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC),
			},
			baseline: &UserBaseline{
				TypicalHours: []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
			},
			expectedSeverity: 0.4,
			anomalyType:      "unusual_time",
		},
		{
			name: "high_velocity_severity",
			authContext: &AuthContext{
				Time: time.Date(2025, 1, 1, 10, 0, 1, 0, time.UTC),
			},
			baseline: &UserBaseline{
				TypicalHours: []int{10},
				LastAuthTime: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			expectedSeverity: 0.9,
			anomalyType:      "high_velocity",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			anomalies, err := analyzer.DetectAnomalies(ctx, tc.authContext, tc.baseline)
			require.NoError(t, err)
			require.NotEmpty(t, anomalies)

			found := false

			for _, anomaly := range anomalies {
				if anomaly.Type == tc.anomalyType {
					require.Equal(t, tc.expectedSeverity, anomaly.Severity)

					found = true

					break
				}
			}

			require.True(t, found, "Expected anomaly type %s not found", tc.anomalyType)
		})
	}
}

func TestAnomalyMetadata(t *testing.T) {
	t.Parallel()

	analyzer := NewDefaultContextAnalyzer(nil, nil)
	ctx := context.Background()

	authContext := &AuthContext{
		Location: &GeoLocation{
			Country: "CN",
			City:    "Beijing",
		},
		Device: &DeviceFingerprint{
			ID: "test-device-id",
		},
		Time: time.Date(2025, 1, 1, 3, 0, 2, 0, time.UTC),
	}

	baseline := &UserBaseline{
		KnownLocations: []GeoLocation{},
		KnownDevices:   []string{},
		TypicalHours:   []int{10, 11, 12},
		LastAuthTime:   time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC),
	}

	anomalies, err := analyzer.DetectAnomalies(ctx, authContext, baseline)
	require.NoError(t, err)

	for _, anomaly := range anomalies {
		require.NotNil(t, anomaly.Metadata)
		require.NotEmpty(t, anomaly.Description)

		switch anomaly.Type {
		case "unknown_location":
			require.Equal(t, "CN", anomaly.Metadata["country"])
			require.Equal(t, "Beijing", anomaly.Metadata["city"])
		case "unknown_device":
			require.Equal(t, "test-device-id", anomaly.Metadata["device_id"])
		case "unusual_time":
			require.Equal(t, 3, anomaly.Metadata["hour"])
		case "high_velocity":
			require.Contains(t, anomaly.Metadata["time_since_last_auth"], "2s")
		}
	}
}
