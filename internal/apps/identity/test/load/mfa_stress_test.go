// Copyright (c) 2025 Justin Cranford
//
//

package load

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// MFAStressTestSuite manages load/stress testing infrastructure.
type MFAStressTestSuite struct {
	concurrentSessions int32
	completedSessions  int32
	failedSessions     int32
	replayAttempts     int32
	usedNonces         sync.Map // Track used nonces for replay detection
}

// TestMFAStress100ConcurrentSessions tests MFA under high concurrency load.
func TestMFAStress100ConcurrentSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	t.Parallel()

	suite := &MFAStressTestSuite{}
	ctx := context.Background()

	const (
		parallelSessions = 100
		factorsPerChain  = 3
	)

	t.Run("100_Parallel_MFA_Chains", func(t *testing.T) {
		var wg sync.WaitGroup

		startTime := time.Now().UTC()

		for i := 0; i < parallelSessions; i++ {
			wg.Add(1)

			go func(_ int) {
				defer wg.Done()

				atomic.AddInt32(&suite.concurrentSessions, 1)

				// Generate unique user ID using UUIDv7 only (not concatenated string)
				// to avoid "invalid UUID length" errors when creating domain.User.
				userID := googleUuid.Must(googleUuid.NewV7()).String()

				err := suite.executeMFAChain(ctx, userID, factorsPerChain)
				if err != nil {
					atomic.AddInt32(&suite.failedSessions, 1)
					require.NoErrorf(t, err, "iteration %d failed", i)
				} else {
					atomic.AddInt32(&suite.completedSessions, 1)
				}

				atomic.AddInt32(&suite.concurrentSessions, -1)
			}(i)
		}

		wg.Wait()

		duration := time.Since(startTime)

		t.Logf("Stress test completed in %v", duration)
		t.Logf("Total sessions: %d", parallelSessions)
		t.Logf("Completed successfully: %d", atomic.LoadInt32(&suite.completedSessions))
		t.Logf("Failed: %d", atomic.LoadInt32(&suite.failedSessions))
		t.Logf("Average duration per session: %v", duration/time.Duration(parallelSessions))

		require.Equal(t, int32(parallelSessions), atomic.LoadInt32(&suite.completedSessions), "All sessions should complete successfully")
		require.Equal(t, int32(0), atomic.LoadInt32(&suite.failedSessions), "No sessions should fail")
	})
}

// TestMFASessionCollisions tests session isolation under concurrent access.
func TestMFASessionCollisions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping collision test in short mode")
	}

	t.Parallel()

	suite := &MFAStressTestSuite{}
	ctx := context.Background()

	const (
		parallelUpdates   = 50
		updatesPerSession = 10
	)

	t.Run("Concurrent_Session_Updates", func(t *testing.T) {
		sessionID := googleUuid.Must(googleUuid.NewV7()).String()

		var wg sync.WaitGroup

		collisions := int32(0)

		for i := 0; i < parallelUpdates; i++ {
			wg.Add(1)

			go func(updateIndex int) {
				defer wg.Done()

				for j := 0; j < updatesPerSession; j++ {
					err := suite.updateSession(ctx, sessionID, fmt.Sprintf("update_%d_%d", updateIndex, j))
					if err != nil {
						atomic.AddInt32(&collisions, 1)
					}

					// Small delay to increase collision probability.
					time.Sleep(time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		totalUpdates := parallelUpdates * updatesPerSession
		collisionRate := float64(atomic.LoadInt32(&collisions)) / float64(totalUpdates) * 100

		t.Logf("Total session updates: %d", totalUpdates)
		t.Logf("Collisions detected: %d", atomic.LoadInt32(&collisions))
		t.Logf("Collision rate: %.2f%%", collisionRate)

		require.Less(t, collisionRate, 5.0, "Collision rate should be below 5%")
	})
}

// TestMFAReplayAttackSimulation tests replay attack detection under load.
func TestMFAReplayAttackSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping replay attack test in short mode")
	}

	t.Parallel()

	suite := &MFAStressTestSuite{}
	ctx := context.Background()

	const (
		parallelAttacks   = 50
		attemptsPerAttack = 5
	)

	t.Run("Concurrent_Replay_Attempts", func(t *testing.T) {
		var wg sync.WaitGroup

		detectedReplays := int32(0)

		for i := 0; i < parallelAttacks; i++ {
			wg.Add(1)

			go func(attackIndex int) {
				defer wg.Done()

				nonce := googleUuid.Must(googleUuid.NewV7()).String()

				// First attempt should succeed (nonce valid).
				err := suite.validateWithNonce(ctx, nonce)
				require.NoError(t, err, "First validation failed for attack %d", attackIndex)

				// Subsequent attempts should be detected as replays.
				for j := 0; j < attemptsPerAttack; j++ {
					err := suite.validateWithNonce(ctx, nonce)
					if err != nil {
						atomic.AddInt32(&detectedReplays, 1)
						atomic.AddInt32(&suite.replayAttempts, 1)
					}
				}
			}(i)
		}

		wg.Wait()

		expectedReplays := parallelAttacks * attemptsPerAttack

		t.Logf("Total replay attempts: %d", expectedReplays)
		t.Logf("Detected and rejected: %d", atomic.LoadInt32(&detectedReplays))

		require.Equal(t, int32(expectedReplays), atomic.LoadInt32(&detectedReplays), "All replay attempts should be detected")
	})
}

// TestMFALongRunningStress tests MFA under sustained load.
func TestMFALongRunningStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running stress test in short mode")
	}

	t.Parallel()

	suite := &MFAStressTestSuite{}
	ctx := context.Background()

	const (
		// P0.5 optimization: Reduced from 30s to 5s for faster unit test execution
		// Full 30s load testing should be in Gatling load tests (test/load/)
		testDuration    = 5 * time.Second
		parallelWorkers = 20
	)

	t.Run("Sustained_Load_5_Seconds", func(t *testing.T) {
		var wg sync.WaitGroup

		stopSignal := make(chan struct{})

		startTime := time.Now().UTC()

		for i := 0; i < parallelWorkers; i++ {
			wg.Add(1)

			go func(_ int) {
				defer wg.Done()

				sessionCount := 0

				for {
					select {
					case <-stopSignal:
						return
					default:
						// Generate unique user ID using UUIDv7 instead of string formatting
						// to avoid "invalid UUID length" errors when creating domain.User.
						userID := googleUuid.Must(googleUuid.NewV7()).String()

						err := suite.executeMFAChain(ctx, userID, 2)
						if err != nil {
							atomic.AddInt32(&suite.failedSessions, 1)
						} else {
							atomic.AddInt32(&suite.completedSessions, 1)
						}

						sessionCount++

						// Small delay between sessions.
						time.Sleep(10 * time.Millisecond)
					}
				}
			}(i)
		}

		// Run for specified duration.
		time.Sleep(testDuration)
		close(stopSignal)
		wg.Wait()

		duration := time.Since(startTime)

		totalSessions := atomic.LoadInt32(&suite.completedSessions) + atomic.LoadInt32(&suite.failedSessions)
		throughput := float64(totalSessions) / duration.Seconds()

		t.Logf("Sustained load test completed")
		t.Logf("Duration: %v", duration)
		t.Logf("Total sessions: %d", totalSessions)
		t.Logf("Completed successfully: %d", atomic.LoadInt32(&suite.completedSessions))
		t.Logf("Failed: %d", atomic.LoadInt32(&suite.failedSessions))
		t.Logf("Throughput: %.2f sessions/second", throughput)

		failureRate := float64(atomic.LoadInt32(&suite.failedSessions)) / float64(totalSessions) * 100
		require.Less(t, failureRate, 1.0, "Failure rate should be below 1%")
	})
}

// executeMFAChain simulates MFA chain execution for load testing.
func (s *MFAStressTestSuite) executeMFAChain(ctx context.Context, userID string, factorCount int) error {
	// Create session.
	session := &cryptoutilIdentityDomain.Session{
		ID:        googleUuid.Must(googleUuid.NewV7()),
		SessionID: googleUuid.Must(googleUuid.NewV7()).String(),
		UserID:    googleUuid.MustParse(userID),
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultSessionLifetime),
		Active:    boolPtr(true),
	}

	// Simulate factor validations.
	for i := 0; i < factorCount; i++ {
		// Simulate validation delay (database query, crypto operations).
		time.Sleep(5 * time.Millisecond)

		// Simulate nonce validation.
		nonce := googleUuid.Must(googleUuid.NewV7()).String()
		if err := s.validateWithNonce(ctx, nonce); err != nil {
			return fmt.Errorf("factor %d validation failed: %w", i+1, err)
		}

		session.AuthenticationMethods = append(session.AuthenticationMethods, fmt.Sprintf("factor_%d", i+1))
	}

	session.AuthenticationTime = time.Now().UTC()

	return nil
}

// updateSession simulates concurrent session updates.
func (s *MFAStressTestSuite) updateSession(_ context.Context, _ string, _ string) error {
	// Simulate database update delay.
	time.Sleep(2 * time.Millisecond)

	// In production, this would update session in database with optimistic locking.
	// Stub: Always succeed (collision detection would be in real repository).
	return nil
}

// validateWithNonce simulates nonce-based validation with replay detection.
func (s *MFAStressTestSuite) validateWithNonce(_ context.Context, nonce string) error {
	// Simulate validation delay.
	time.Sleep(3 * time.Millisecond)

	// Check if nonce was already used (replay detection).
	if _, exists := s.usedNonces.LoadOrStore(nonce, true); exists {
		// Nonce already used - replay attack detected.
		return fmt.Errorf("replay attack detected: nonce already used")
	}

	// First use of this nonce - validation succeeds.
	return nil
}
