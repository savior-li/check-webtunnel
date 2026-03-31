package validator

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"tor-bridge-collector/pkg/bridge"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator(10, 5)
	assert.NotNil(t, v)
	assert.Equal(t, 10*time.Second, v.timeout)
	assert.Equal(t, 5, v.workers)
}

func TestValidator_Validate_Timeout(t *testing.T) {
	v := NewValidator(1, 1)

	bridge := &bridge.Bridge{
		ID:      1,
		Address: "10.255.255.1",
		Port:    12345,
	}

	result, err := v.Validate(bridge)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsAvailable)
	assert.Equal(t, int64(1), result.BridgeID)
}

func TestValidator_Validate_ConcurrentAccess(t *testing.T) {
	v := NewValidator(5, 10)

	bridges := make([]bridge.Bridge, 20)
	for i := range bridges {
		bridges[i] = bridge.Bridge{
			ID:      int64(i),
			Address: "192.0.2.1",
			Port:    12345,
		}
	}

	var mu sync.Mutex
	results := make([]*ValidationResult, 0)

	err := v.ValidateConcurrent(bridges, func(result *ValidationResult) {
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	})

	assert.NoError(t, err)
	assert.Len(t, results, 20)
}

func TestValidator_ValidateConcurrent_Empty(t *testing.T) {
	v := NewValidator(5, 5)

	err := v.ValidateConcurrent([]bridge.Bridge{}, func(result *ValidationResult) {
		t.Error("Callback should not be called for empty slice")
	})

	assert.NoError(t, err)
}

func TestValidator_ValidateConcurrent_Callback(t *testing.T) {
	v := NewValidator(5, 3)

	bridges := []bridge.Bridge{
		{ID: 1, Address: "192.0.2.1", Port: 12345},
		{ID: 2, Address: "192.0.2.2", Port: 12345},
	}

	callCount := 0
	var mu sync.Mutex

	err := v.ValidateConcurrent(bridges, func(result *ValidationResult) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestValidationResult_Fields(t *testing.T) {
	result := &ValidationResult{
		BridgeID:     123,
		IsAvailable:  true,
		ResponseTime: 100,
		Error:        nil,
		ValidatedAt:  time.Now(),
	}

	assert.Equal(t, int64(123), result.BridgeID)
	assert.True(t, result.IsAvailable)
	assert.Equal(t, 100, result.ResponseTime)
	assert.Nil(t, result.Error)
	assert.False(t, result.ValidatedAt.IsZero())
}

func TestValidator_ValidateAll(t *testing.T) {
	v := NewValidator(5, 2)

	bridges := []bridge.Bridge{
		{ID: 1, Address: "192.0.2.1", Port: 12345},
		{ID: 2, Address: "192.0.2.2", Port: 12345},
	}

	results, err := v.ValidateAll(bridges)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestValidator_ValidateAll_Empty(t *testing.T) {
	v := NewValidator(5, 5)

	results, err := v.ValidateAll([]bridge.Bridge{})

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestValidator_Validate_InvalidAddress(t *testing.T) {
	v := NewValidator(1, 1)

	bridge := &bridge.Bridge{ID: 1, Address: "invalid-address", Port: 9999}

	result, err := v.Validate(bridge)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsAvailable)
}
