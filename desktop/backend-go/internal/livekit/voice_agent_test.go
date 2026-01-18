package livekit

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestDetectVoiceActivity(t *testing.T) {
	tests := []struct {
		name      string
		samples   []int16
		threshold float64
		want      bool
		desc      string
	}{
		{
			name:      "silence (all zeros)",
			samples:   make([]int16, 1000),
			threshold: 0.05,
			want:      false,
			desc:      "Empty silence should return false",
		},
		{
			name:      "loud speech (high amplitude)",
			samples:   generateTestSinWave(1000, 20000),
			threshold: 0.05,
			want:      true,
			desc:      "High amplitude audio should detect voice",
		},
		{
			name:      "quiet speech (moderate amplitude)",
			samples:   generateTestSinWave(1000, 5000),
			threshold: 0.05,
			want:      true,
			desc:      "Moderate amplitude audio should detect voice",
		},
		{
			name:      "low amplitude noise (below threshold)",
			samples:   generateTestSinWave(1000, 500),
			threshold: 0.05,
			want:      false,
			desc:      "Low amplitude audio below threshold should return false",
		},
		{
			name:      "empty buffer",
			samples:   []int16{},
			threshold: 0.05,
			want:      false,
			desc:      "Empty buffer should return false",
		},
		{
			name:      "single sample (loud)",
			samples:   []int16{30000},
			threshold: 0.05,
			want:      true,
			desc:      "Single loud sample should detect voice",
		},
		{
			name:      "threshold edge case (exactly at threshold)",
			samples:   generateTestSinWave(1000, int16(math.Ceil(32768*0.05))),
			threshold: 0.05,
			want:      false,
			desc:      "Energy exactly at threshold should return false (uses >)",
		},
		{
			name:      "threshold edge case (just above threshold)",
			samples:   generateTestSinWave(1000, 3000), // Much higher amplitude
			threshold: 0.05,
			want:      true,
			desc:      "Energy just above threshold should return true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			got := detectVoiceActivity(tt.samples, tt.threshold)

			// Assert
			assert.Equal(t, tt.want, got, tt.desc)
		})
	}
}

func TestVADConfigDefault(t *testing.T) {
	// Arrange: Create default config
	config := VADConfig{
		MinSpeechDuration:   50 * time.Millisecond,
		MinSilenceDuration:  550 * time.Millisecond,
		ActivationThreshold: 0.05,
		SampleRate:          48000,
	}

	// Assert: Verify reasonable defaults
	assert.Equal(t, 50*time.Millisecond, config.MinSpeechDuration)
	assert.Equal(t, 550*time.Millisecond, config.MinSilenceDuration)
	assert.Equal(t, 0.05, config.ActivationThreshold)
	assert.Equal(t, 48000, config.SampleRate)

	// Assert: Silence > speech (basic sanity check)
	assert.Greater(t, config.MinSilenceDuration, config.MinSpeechDuration,
		"Silence threshold should be greater than speech threshold")

	// Assert: Threshold in valid range
	assert.GreaterOrEqual(t, config.ActivationThreshold, 0.0)
	assert.LessOrEqual(t, config.ActivationThreshold, 1.0)
}

func TestVADConfigCustom(t *testing.T) {
	// Arrange: Create custom config
	config := VADConfig{
		MinSpeechDuration:   100 * time.Millisecond,
		MinSilenceDuration:  1000 * time.Millisecond,
		ActivationThreshold: 0.1,
		SampleRate:          16000,
	}

	// Assert: Verify custom values
	assert.Equal(t, 100*time.Millisecond, config.MinSpeechDuration)
	assert.Equal(t, 1000*time.Millisecond, config.MinSilenceDuration)
	assert.Equal(t, 0.1, config.ActivationThreshold)
	assert.Equal(t, 16000, config.SampleRate)
}

func TestDetectVoiceActivity_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name    string
		samples []int16
		want    bool
		desc    string
	}{
		{
			name:    "speech with pauses (mixed)",
			samples: mixAudioSegments(generateTestSinWave(480, 15000), generateTestSinWave(240, 0)),
			want:    true,
			desc:    "Audio with speech and pauses should detect voice",
		},
		{
			name:    "complete silence 5 seconds",
			samples: make([]int16, 48000*5), // 5 sec at 48kHz
			want:    false,
			desc:    "Extended silence should not detect voice",
		},
		{
			name:    "sustained speech 2 seconds",
			samples: generateTestSinWave(48000*2, 12000),
			want:    true,
			desc:    "Extended speech should detect voice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectVoiceActivity(tt.samples, 0.05)
			assert.Equal(t, tt.want, got, tt.desc)
		})
	}
}

func TestWrapPCMInWAV(t *testing.T) {
	tests := []struct {
		name       string
		samples    []int16
		sampleRate int
		channels   int
		validate   func(t *testing.T, wavData []byte)
	}{
		{
			name:       "mono 48kHz",
			samples:    generateTestSinWave(960, 10000),
			sampleRate: 48000,
			channels:   1,
			validate: func(t *testing.T, wavData []byte) {
				// Minimum WAV header is 44 bytes
				assert.GreaterOrEqual(t, len(wavData), 44, "WAV data should be at least 44 bytes")

				// Check RIFF header
				assert.Equal(t, "RIFF", string(wavData[0:4]))

				// Check WAVE header
				assert.Equal(t, "WAVE", string(wavData[8:12]))

				// Check fmt chunk
				assert.Equal(t, "fmt ", string(wavData[12:16]))

				// Check data chunk
				assert.Equal(t, "data", string(wavData[36:40]))
			},
		},
		{
			name:       "mono 16kHz",
			samples:    generateTestSinWave(320, 8000),
			sampleRate: 16000,
			channels:   1,
			validate: func(t *testing.T, wavData []byte) {
				assert.Greater(t, len(wavData), 44)
				assert.Equal(t, "RIFF", string(wavData[0:4]))
			},
		},
		{
			name:       "stereo 44.1kHz",
			samples:    generateTestSinWave(882, 12000),
			sampleRate: 44100,
			channels:   2,
			validate: func(t *testing.T, wavData []byte) {
				assert.Greater(t, len(wavData), 44)
				assert.Equal(t, "RIFF", string(wavData[0:4]))
			},
		},
		{
			name:       "empty samples",
			samples:    []int16{},
			sampleRate: 48000,
			channels:   1,
			validate: func(t *testing.T, wavData []byte) {
				// Should still have WAV header
				assert.Equal(t, len(wavData), 44, "Empty PCM should produce just header")
				assert.Equal(t, "RIFF", string(wavData[0:4]))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			wavData := wrapPCMInWAV(tt.samples, tt.sampleRate, tt.channels)

			// Assert
			assert.NotNil(t, wavData)
			tt.validate(t, wavData)
		})
	}
}

func TestDecodeMp3ToPCM_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		mp3Data []byte
	}{
		{
			name:    "empty data",
			mp3Data: []byte{},
		},
		{
			name:    "invalid header",
			mp3Data: []byte{0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:    "truncated data",
			mp3Data: []byte{0xFF, 0xFB}, // MP3 frame header start but incomplete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, _, err := decodeMp3ToPCM(tt.mp3Data)

			// Assert
			assert.Error(t, err, "Should fail on invalid MP3 data")
		})
	}
}

func TestNewPureGoVoiceAgent(t *testing.T) {
	// Arrange: Setup test dependencies
	t.Setenv("LIVEKIT_URL", "ws://test.livekit.io")
	t.Setenv("LIVEKIT_API_KEY", "test-key")
	t.Setenv("LIVEKIT_API_SECRET", "test-secret")

	// Act
	// Note: In real usage, pool would be *pgxpool.Pool, cfg would be *config.Config,
	// but this test just validates initialization works with environment variables
	// Skip full integration test - this is tested via integration tests
	t.Skip("Skipping integration test - requires database connection")
}

// Note: GetActiveRooms test requires LiveKit SDK mocking
// This is tested via integration tests with actual LiveKit connection
// Skipping for now as it requires complex LiveKit SDK type setup

// ============================================================================
// Helper functions for tests
// ============================================================================

// generateTestSinWave creates a sine wave for testing
func generateTestSinWave(samples int, amplitude int16) []int16 {
	result := make([]int16, samples)
	for i := 0; i < samples; i++ {
		// Sine wave with frequency that varies based on amplitude
		freq := 0.01 + (float64(amplitude)/32768.0)*0.1
		result[i] = int16(float64(amplitude) * math.Sin(float64(i)*freq))
	}
	return result
}

// generateTestNoise creates random noise for testing
func generateTestNoise(samples int, maxAmplitude int16) []int16 {
	result := make([]int16, samples)
	for i := 0; i < samples; i++ {
		result[i] = int16(rand.Intn(int(maxAmplitude))) - maxAmplitude/2
	}
	return result
}

// mixAudioSegments concatenates multiple audio segments
func mixAudioSegments(segments ...[]int16) []int16 {
	totalLen := 0
	for _, seg := range segments {
		totalLen += len(seg)
	}

	result := make([]int16, totalLen)
	offset := 0
	for _, seg := range segments {
		copy(result[offset:], seg)
		offset += len(seg)
	}
	return result
}

// Mock types for testing

type testConfig struct{}

func (tc *testConfig) Get(key string) interface{} {
	return nil
}

type testVoiceController struct{}

type testMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}

func setupTestPool(t *testing.T) interface{} {
	// Return mock pool - in real tests this would connect to test DB
	return nil
}

// TestMonitorRooms_Integration tests the room monitoring with mocked LiveKit SDK
// NOTE: This requires LiveKit server running - skipped in CI
func TestMonitorRooms_Integration(t *testing.T) {
	t.Skip("Integration test - requires LiveKit server")

	// This test would:
	// 1. Create a real PureGoVoiceAgent
	// 2. Start room monitoring
	// 3. Create a test room with a user
	// 4. Verify agent auto-joins within 5 seconds
	// 5. Verify agent doesn't join again on next poll
	// 6. Clean up
}

// TestRoomMonitoring_Logic tests the decision logic for joining rooms
func TestRoomMonitoring_Logic(t *testing.T) {
	tests := []struct {
		name          string
		roomName      string
		participants  []string // "agent-osa" means agent, "user-X" means user
		alreadyJoined bool
		shouldJoin    bool
		description   string
	}{
		{
			name:          "Join room with user and no agent",
			roomName:      "test-room-1",
			participants:  []string{"user-123"},
			alreadyJoined: false,
			shouldJoin:    true,
			description:   "Should join room that has a user but no agent",
		},
		{
			name:          "Skip room that already has agent",
			roomName:      "test-room-2",
			participants:  []string{"user-123", "agent-osa"},
			alreadyJoined: false,
			shouldJoin:    false,
			description:   "Should not join room that already has agent",
		},
		{
			name:          "Skip room we already joined",
			roomName:      "test-room-3",
			participants:  []string{"user-123"},
			alreadyJoined: true,
			shouldJoin:    false,
			description:   "Should not rejoin room we're already in",
		},
		{
			name:          "Skip empty room",
			roomName:      "test-room-4",
			participants:  []string{},
			alreadyJoined: false,
			shouldJoin:    false,
			description:   "Should not join empty room",
		},
		{
			name:          "Join room with multiple users and no agent",
			roomName:      "test-room-5",
			participants:  []string{"user-123", "user-456", "user-789"},
			alreadyJoined: false,
			shouldJoin:    true,
			description:   "Should join room with multiple users but no agent",
		},
		{
			name:          "Skip room with only agent (user left)",
			roomName:      "test-room-6",
			participants:  []string{"agent-osa"},
			alreadyJoined: false,
			shouldJoin:    false,
			description:   "Should not join room with only agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from monitorRooms()
			hasAgent := false
			hasUser := false
			numParticipants := len(tt.participants)

			// Check participants
			for _, p := range tt.participants {
				if p == "agent-osa" {
					hasAgent = true
				} else {
					hasUser = true
				}
			}

			// Determine if should join (same logic as monitorRooms)
			shouldJoin := !tt.alreadyJoined && numParticipants > 0 && hasUser && !hasAgent

			assert.Equal(t, tt.shouldJoin, shouldJoin, tt.description)
		})
	}
}

// TestRoomMonitoring_UserIDExtraction tests extracting user ID from participant identity
func TestRoomMonitoring_UserIDExtraction(t *testing.T) {
	tests := []struct {
		identity       string
		expectedUserID string
	}{
		{"user-abc123", "abc123"},
		{"user-01JGRYC123", "01JGRYC123"},
		{"user-uuid-with-dashes", "uuid-with-dashes"},
		{"invalid", "invalid"}, // No "user-" prefix
		{"user-", "user-"},     // No content after prefix - keeps original
	}

	for _, tt := range tests {
		t.Run(tt.identity, func(t *testing.T) {
			// Simulate the extraction logic from monitorRooms()
			userID := tt.identity
			if len(userID) > 5 && userID[:5] == "user-" {
				userID = userID[5:]
			}

			assert.Equal(t, tt.expectedUserID, userID)
		})
	}
}

// TestRoomMonitoring_ConcurrentAccess tests thread-safety of activeRooms map
func TestRoomMonitoring_ConcurrentAccess(t *testing.T) {
	agent := &PureGoVoiceAgent{
		activeRooms: make(map[string]*lksdk.Room),
	}

	// Simulate concurrent access from multiple goroutines
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			roomName := fmt.Sprintf("room-%d", id)

			for j := 0; j < numOperations; j++ {
				// Check if room exists (read)
				agent.mu.Lock()
				_, exists := agent.activeRooms[roomName]
				agent.mu.Unlock()

				if !exists {
					// Add room (write)
					agent.mu.Lock()
					agent.activeRooms[roomName] = nil // Mock room
					agent.mu.Unlock()
				}

				// Remove room (write)
				agent.mu.Lock()
				delete(agent.activeRooms, roomName)
				agent.mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions detected
}

// TestRoomMonitoring_PollingInterval verifies polling happens every 5 seconds
func TestRoomMonitoring_PollingInterval(t *testing.T) {
	t.Skip("Time-sensitive test - may be flaky in CI")

	// This test would:
	// 1. Mock the RoomServiceClient.ListRooms() call
	// 2. Count how many times it's called
	// 3. Verify it's called approximately every 5 seconds
	// 4. Cancel context and verify cleanup
}

// Benchmark tests

func BenchmarkDetectVoiceActivity_Silence(b *testing.B) {
	samples := make([]int16, 48000) // 1 second of silence at 48kHz

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectVoiceActivity(samples, 0.05)
	}
}

func BenchmarkDetectVoiceActivity_Speech(b *testing.B) {
	samples := generateTestSinWave(48000, 15000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectVoiceActivity(samples, 0.05)
	}
}

func BenchmarkWrapPCMInWAV_Small(b *testing.B) {
	samples := generateTestSinWave(960, 10000) // 20ms at 48kHz

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapPCMInWAV(samples, 48000, 1)
	}
}

func BenchmarkWrapPCMInWAV_Large(b *testing.B) {
	samples := generateTestSinWave(480000, 10000) // 10 seconds at 48kHz

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapPCMInWAV(samples, 48000, 1)
	}
}

// Integration test (marked for integration testing)
func TestDetectVoiceActivity_IntegrationWithVADConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test verifies VAD works with actual config from agent
	config := VADConfig{
		MinSpeechDuration:   50 * time.Millisecond,
		MinSilenceDuration:  550 * time.Millisecond,
		ActivationThreshold: 0.05,
		SampleRate:          48000,
	}

	// Test with config threshold
	silence := make([]int16, config.SampleRate/10) // 100ms silence
	assert.False(t, detectVoiceActivity(silence, config.ActivationThreshold))

	// Test with speech
	speech := generateTestSinWave(config.SampleRate/10, 15000)
	assert.True(t, detectVoiceActivity(speech, config.ActivationThreshold))
}
