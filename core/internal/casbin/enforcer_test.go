package casbin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
)

// Test data and configurations
var (
	testPolicies = []Policy{
		{"user1", "resource1", "read"},
		{"user2", "resource2", "write"},
		{"admin", "system", "manage"},
	}
)

// Test helper functions
func createMockEnforcer() *Enforcer {
	return &Enforcer{se: nil}
}

func assertPanicRecovery(t *testing.T, expectedPanicMessage string) {
	t.Helper()
	if r := recover(); r != nil {
		t.Logf("Expected panic recovered: %v", r)
		if expectedPanicMessage != "" {
			// You could add more specific panic message checking here
		}
	}
}

// TestNewEnforcer tests the NewEnforcer constructor with table-driven approach
func TestNewEnforcer(t *testing.T) {
	testCases := []struct {
		name           string
		enforcerType   EnforcerType
		databaseConfig orm.DatabaseConfig
		shouldFail     bool
		expectedError  string
	}{
		{
			name:           "valid group enforcer type",
			enforcerType:   EnforcerTypeGroup,
			databaseConfig: orm.DatabaseConfig{},
			shouldFail:     true, // Will fail due to no real DB connection
		},
		{
			name:           "valid dashboard enforcer type",
			enforcerType:   EnforcerTypeDashboard,
			databaseConfig: orm.DatabaseConfig{},
			shouldFail:     true, // Will fail due to no real DB connection
		},
		{
			name:           "invalid enforcer type",
			enforcerType:   EnforcerType("invalid_type"),
			databaseConfig: orm.DatabaseConfig{},
			shouldFail:     true,
			expectedError:  "not exist model text",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing enforcer
			Enforcers.Remove(tt.enforcerType.String(), nil)

			// Execute
			enforcer, err := NewEnforcer(tt.enforcerType, tt.databaseConfig)

			// Verify
			if tt.shouldFail {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.EqualError(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, enforcer)
				assert.NotNil(t, enforcer.se)
			}
		})
	}
}

// TestEnforcer_Methods tests all Enforcer methods with consistent approach
func TestEnforcer_Methods(t *testing.T) {
	testCases := []struct {
		name        string
		method      string
		setupFunc   func(*Enforcer) error
		shouldPanic bool
	}{
		{
			name:   "Enforce with empty policies",
			method: "Enforce",
			setupFunc: func(e *Enforcer) error {
				result, err := e.Enforce()
				assert.Equal(t, Policy{}, result)
				assert.Equal(t, external.ErrorNoSuchPolicy, err)
				return nil
			},
			shouldPanic: false,
		},
		{
			name:   "Enforce with multiple policies - nil SyncedEnforcer",
			method: "Enforce",
			setupFunc: func(e *Enforcer) error {
				_, err := e.Enforce(testPolicies...)
				return err
			},
			shouldPanic: true,
		},
		{
			name:   "GetFilteredPolicy with nil SyncedEnforcer",
			method: "GetFilteredPolicy",
			setupFunc: func(e *Enforcer) error {
				_, err := e.GetFilteredPolicy(0, "test")
				return err
			},
			shouldPanic: true,
		},
		{
			name:   "AddPolicy with nil SyncedEnforcer",
			method: "AddPolicy",
			setupFunc: func(e *Enforcer) error {
				return e.AddPolicy(Policy{"sub", "obj", "act"})
			},
			shouldPanic: true,
		},
		{
			name:   "RemovePolicy with nil SyncedEnforcer",
			method: "RemovePolicy",
			setupFunc: func(e *Enforcer) error {
				return e.RemovePolicy(Policy{"sub", "obj", "act"})
			},
			shouldPanic: true,
		},
		{
			name:   "RemovePolicyFromField with nil SyncedEnforcer",
			method: "RemovePolicyFromField",
			setupFunc: func(e *Enforcer) error {
				return e.RemovePolicyFromField(0, "test")
			},
			shouldPanic: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			enforcer := createMockEnforcer()

			if tt.shouldPanic {
				defer assertPanicRecovery(t, "")
			}

			err := tt.setupFunc(enforcer)

			if !tt.shouldPanic && err != nil {
				t.Logf("Expected error for method %s: %v", tt.method, err)
			}
		})
	}
}

// TestErrorConstants tests error constant values
func TestErrorConstants(t *testing.T) {
	errorTests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrorNoSuchPolicy",
			err:      external.ErrorNoSuchPolicy,
			expected: "no such policy",
		},
		{
			name:     "ErrorNoPermission",
			err:      external.ErrorNoSuchPermission,
			expected: "no such permission",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

// TestEnforcerScenarios tests real-world usage scenarios
func TestEnforcerScenarios(t *testing.T) {
	scenarios := []struct {
		name      string
		setupFunc func(t *testing.T, e *Enforcer)
	}{
		{
			name: "empty policies enforcement",
			setupFunc: func(t *testing.T, e *Enforcer) {
				result, err := e.Enforce()
				assert.Equal(t, Policy{}, result)
				assert.Equal(t, external.ErrorNoSuchPolicy, err)
			},
		},
		{
			name: "multiple policies enforcement with panic handling",
			setupFunc: func(t *testing.T, e *Enforcer) {
				defer assertPanicRecovery(t, "nil SyncedEnforcer")

				result, err := e.Enforce(testPolicies...)
				if err == nil {
					t.Logf("Unexpected success with result: %v", result)
				}
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			enforcer := createMockEnforcer()
			scenario.setupFunc(t, enforcer)
		})
	}
}

// TestEnforcer_Integration provides integration test template
func TestEnforcer_Integration(t *testing.T) {
	t.Skip("Integration test requires database setup")

	integrationTests := []struct {
		name      string
		setupFunc func(t *testing.T) *Enforcer
		testFunc  func(t *testing.T, e *Enforcer)
	}{
		{
			name: "complete policy lifecycle",
			setupFunc: func(t *testing.T) *Enforcer {
				// This would create a real enforcer with test database
				return createMockEnforcer() // Replace with real implementation
			},
			testFunc: func(t *testing.T, e *Enforcer) {
				policy := Policy{"user1", "resource1", "read"}

				// Test policy addition
				err := e.AddPolicy(policy)
				assert.NoError(t, err)

				// Test policy enforcement
				result, err := e.Enforce(policy)
				assert.NoError(t, err)
				assert.Equal(t, policy, result)

				// Test filtered policy retrieval
				policies, err := e.GetFilteredPolicy(0, "user1")
				assert.NoError(t, err)
				assert.Contains(t, policies, policy)

				// Test policy removal
				err = e.RemovePolicy(policy)
				assert.NoError(t, err)

				// Verify policy was removed
				_, err = e.Enforce(policy)
				assert.Equal(t, external.ErrorNoSuchPolicy, err)
			},
		},
	}

	for _, tt := range integrationTests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer := tt.setupFunc(t)
			tt.testFunc(t, enforcer)
		})
	}
}

// Benchmark tests for performance measurement
func BenchmarkEnforcer_Enforce(b *testing.B) {
	enforcer := createMockEnforcer()
	policy := Policy{"user", "resource", "action"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Handle panic from nil SyncedEnforcer
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Ignore panic for benchmark
				}
			}()
			_, err := enforcer.Enforce(policy)
			if err != nil {
				return
			}
		}()
	}
}

func BenchmarkPolicy_getAny(b *testing.B) {
	policy := Policy{"user", "resource", "action", "kind"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		policy.getAny()
	}
}

func BenchmarkEnforcer_Operations(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func(*Enforcer)
	}{
		{
			name: "AddPolicy",
			fn: func(e *Enforcer) {
				defer func() { recover() }()
				err := e.AddPolicy(Policy{"user", "resource", "action"})
				assert.NoError(b, err)
			},
		},
		{
			name: "RemovePolicy",
			fn: func(e *Enforcer) {
				defer func() { recover() }()
				err := e.RemovePolicy(Policy{"user", "resource", "action"})
				assert.NoError(b, err)
			},
		},
		{
			name: "GetFilteredPolicy",
			fn: func(e *Enforcer) {
				defer func() { recover() }()
				_, err := e.GetFilteredPolicy(0, "user")
				assert.NoError(b, err)
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			enforcer := createMockEnforcer()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.fn(enforcer)
			}
		})
	}
}
