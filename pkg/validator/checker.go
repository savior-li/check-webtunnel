package validator

import (
	"fmt"
	"net"
	"sync"
	"time"

	"tor-bridge-collector/pkg/bridge"
)

type ValidationResult struct {
	BridgeID     int64
	IsAvailable  bool
	ResponseTime int
	Error        error
	ValidatedAt  time.Time
}

type Validator struct {
	timeout time.Duration
	workers int
}

func NewValidator(timeout int, workers int) *Validator {
	return &Validator{
		timeout: time.Duration(timeout) * time.Second,
		workers: workers,
	}
}

func (v *Validator) Validate(b *bridge.Bridge) (*ValidationResult, error) {
	result := &ValidationResult{
		BridgeID:    b.ID,
		IsAvailable: false,
		ValidatedAt: time.Now(),
	}

	start := time.Now()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", b.Address, b.Port), v.timeout)
	if err != nil {
		result.Error = err
		result.IsAvailable = false
		return result, nil
	}
	defer conn.Close()

	result.ResponseTime = int(time.Since(start).Milliseconds())
	result.IsAvailable = true
	return result, nil
}

func (v *Validator) ValidateAll(bridges []bridge.Bridge) ([]ValidationResult, error) {
	if len(bridges) == 0 {
		return []ValidationResult{}, nil
	}

	results := make([]ValidationResult, len(bridges))
	sem := make(chan struct{}, v.workers)
	var wg sync.WaitGroup

	for i, b := range bridges {
		wg.Add(1)
		go func(i int, b bridge.Bridge) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, _ := v.Validate(&b)
			results[i] = *result
		}(i, b)
	}

	wg.Wait()
	return results, nil
}

func (v *Validator) ValidateConcurrent(bridges []bridge.Bridge, callback func(*ValidationResult)) error {
	if len(bridges) == 0 {
		return nil
	}

	sem := make(chan struct{}, v.workers)
	var wg sync.WaitGroup

	for _, b := range bridges {
		wg.Add(1)
		go func(b bridge.Bridge) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, _ := v.Validate(&b)
			if callback != nil {
				callback(result)
			}
		}(b)
	}

	wg.Wait()
	return nil
}
