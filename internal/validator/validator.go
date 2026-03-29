package validator

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"tor-bridge-collector/internal/storage"
	"tor-bridge-collector/pkg/models"
)

type Validator struct {
	storage     *storage.Storage
	timeout     time.Duration
	concurrency int
	retry       int
}

type ValidationResult struct {
	Bridge      *models.Bridge
	IsReachable bool
	Latency     float64
	ErrorMsg    string
}

func New(s *storage.Storage, timeout, concurrency, retry int) *Validator {
	return &Validator{
		storage:     s,
		timeout:     time.Duration(timeout) * time.Second,
		concurrency: concurrency,
		retry:       retry,
	}
}

func (v *Validator) ValidateBridge(ctx context.Context, bridge *models.Bridge) *ValidationResult {
	result := &ValidationResult{Bridge: bridge}

	var conn net.Conn
	var err error
	var latency float64

	for i := 0; i <= v.retry; i++ {
		start := time.Now()
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", bridge.Address, bridge.Port), v.timeout)
		latency = time.Since(start).Seconds() * 1000

		if err == nil {
			conn.Close()
			result.IsReachable = true
			result.Latency = latency
			return result
		}

		if i < v.retry {
			time.Sleep(time.Duration(2<<uint(i)) * time.Second)
		}
	}

	result.IsReachable = false
	result.Latency = latency
	result.ErrorMsg = err.Error()
	return result
}

func (v *Validator) ValidateBridges(ctx context.Context, bridges []models.Bridge) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, len(bridges))
	resultChan := make(chan *ValidationResult, len(bridges))
	bridgeChan := make(chan models.Bridge, len(bridges))

	var wg sync.WaitGroup
	sem := make(chan struct{}, v.concurrency)

	for _, b := range bridges {
		bridgeChan <- b
	}
	close(bridgeChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	go func() {
		for bridge := range bridgeChan {
			wg.Add(1)
			go func(b models.Bridge) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				result := v.ValidateBridge(ctx, &b)
				resultChan <- result
			}(bridge)
		}
	}()

	for result := range resultChan {
		results = append(results, *result)
	}

	return results, nil
}

func (v *Validator) ValidateAndSave(ctx context.Context, bridges []models.Bridge) (int, int, error) {
	results, err := v.ValidateBridges(ctx, bridges)
	if err != nil {
		return 0, 0, err
	}

	var validCount, invalidCount int32

	var wg sync.WaitGroup
	for _, result := range results {
		wg.Add(1)
		go func(r ValidationResult) {
			defer wg.Done()

			bridge := r.Bridge
			bridge.IsValid = r.IsReachable

			if r.IsReachable {
				bridge.AvgLatency = r.Latency
				bridge.SuccessRate = 100.0
				atomic.AddInt32(&validCount, 1)
			} else {
				atomic.AddInt32(&invalidCount, 1)
			}

			if err := v.storage.UpdateBridge(bridge); err != nil {
				return
			}

			history := &models.ValidationHistory{
				BridgeID:    bridge.ID,
				TestedAt:    time.Now(),
				Latency:     r.Latency,
				IsReachable: r.IsReachable,
				ErrorMsg:    r.ErrorMsg,
			}
			v.storage.CreateValidationHistory(history)
		}(result)
	}
	wg.Wait()

	return int(validCount), int(invalidCount), nil
}

func (v *Validator) UpdateBridgeStats(bridge *models.Bridge) error {
	var totalLatency float64
	var totalTests, successfulTests int64

	var histories []models.ValidationHistory
	histories, err := v.storage.GetValidationHistory(bridge.ID, 100)
	if err != nil {
		return err
	}

	for _, h := range histories {
		totalTests++
		if h.IsReachable {
			successfulTests++
			totalLatency += h.Latency
		}
	}

	if totalTests > 0 {
		bridge.AvgLatency = totalLatency / float64(successfulTests)
		bridge.SuccessRate = float64(successfulTests) / float64(totalTests) * 100
	}

	return v.storage.UpdateBridge(bridge)
}
