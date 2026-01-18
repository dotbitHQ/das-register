package internal

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

const (
	DefaultMaxRetries   = 3
	DefaultMaxBackoff   = 30 * time.Second
	DefaultInitialDelay = 1 * time.Second
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxBackoff   time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   DefaultMaxRetries,
		InitialDelay: DefaultInitialDelay,
		MaxBackoff:   DefaultMaxBackoff,
	}
}

// calculateBackoff calculates exponential backoff duration
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	backoff := float64(config.InitialDelay) * math.Pow(2, float64(attempt))
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}
	return time.Duration(backoff)
}

// GetTipBlockNumberWithRetry wraps GetTipBlockNumber with retry logic
func GetTipBlockNumberWithRetry(ctx context.Context, client rpc.Client, config RetryConfig) (uint64, error) {
	var lastErr error
	for i := 0; i < config.MaxRetries; i++ {
		blockNumber, err := client.GetTipBlockNumber(ctx)
		if err == nil {
			return blockNumber, nil
		}
		lastErr = err
		backoff := calculateBackoff(i, config)
		log.Warnf("GetTipBlockNumber retry %d/%d err: %v, backoff: %v", i+1, config.MaxRetries, err, backoff)

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return 0, fmt.Errorf("GetTipBlockNumber failed after %d retries: %v", config.MaxRetries, lastErr)
}

// GetBlockByNumberWithRetry wraps GetBlockByNumber with retry logic
func GetBlockByNumberWithRetry(ctx context.Context, client rpc.Client, blockNumber uint64, config RetryConfig) (*types.Block, error) {
	var lastErr error
	for i := 0; i < config.MaxRetries; i++ {
		block, err := client.GetBlockByNumber(ctx, blockNumber)
		if err == nil {
			return block, nil
		}
		lastErr = err
		backoff := calculateBackoff(i, config)
		log.Warnf("GetBlockByNumber(%d) retry %d/%d err: %v, backoff: %v", blockNumber, i+1, config.MaxRetries, err, backoff)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, fmt.Errorf("GetBlockByNumber(%d) failed after %d retries: %v", blockNumber, config.MaxRetries, lastErr)
}

// GetTransactionWithRetry wraps GetTransaction with retry logic
func GetTransactionWithRetry(ctx context.Context, client rpc.Client, hash types.Hash, config RetryConfig) (*types.TransactionWithStatus, error) {
	var lastErr error
	for i := 0; i < config.MaxRetries; i++ {
		tx, err := client.GetTransaction(ctx, hash)
		if err == nil {
			return tx, nil
		}
		lastErr = err
		backoff := calculateBackoff(i, config)
		log.Warnf("GetTransaction(%s) retry %d/%d err: %v, backoff: %v", hash.String(), i+1, config.MaxRetries, err, backoff)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, fmt.Errorf("GetTransaction(%s) failed after %d retries: %v", hash.String(), config.MaxRetries, lastErr)
}
