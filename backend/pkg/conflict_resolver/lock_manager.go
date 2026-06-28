package conflict_resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/aiops/AiOpsHub/backend/pkg/redis"
)

type LockManager struct {
	RedisClient *redis.RedisClient
}

func NewLockManager(redisClient *redis.RedisClient) *LockManager {
	manager := &LockManager{
		RedisClient: redisClient,
	}

	logger.Info("Created Lock Manager")
	return manager
}

func (lm *LockManager) AcquireLock(ctx context.Context, resourceID, agentID string, timeout time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", resourceID)

	acquired, err := lm.RedisClient.SetNX(ctx, lockKey, agentID, timeout)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to acquire lock: %v", err))
		return false, err
	}

	if acquired {
		logger.Info(fmt.Sprintf("Agent %s acquired lock for resource %s (timeout: %v)", agentID, resourceID, timeout))
		return true, nil
	}

	logger.Info(fmt.Sprintf("Agent %s failed to acquire lock for resource %s (resource busy)", agentID, resourceID))
	return false, nil
}

func (lm *LockManager) ReleaseLock(ctx context.Context, resourceID, agentID string) error {
	lockKey := fmt.Sprintf("lock:%s", resourceID)

	currentHolder, err := lm.RedisClient.Get(ctx, lockKey)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get lock holder: %v", err))
		return err
	}

	if currentHolder != agentID {
		logger.Error(fmt.Sprintf("Agent %s is not the lock holder for resource %s (current holder: %s)",
			agentID, resourceID, currentHolder))
		return fmt.Errorf("not the lock holder")
	}

	err = lm.RedisClient.Del(ctx, lockKey)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to release lock: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Agent %s released lock for resource %s", agentID, resourceID))
	return nil
}

func (lm *LockManager) WaitLock(ctx context.Context, resourceID, agentID string, maxWait time.Duration) (bool, error) {
	startTime := time.Now()
	for {
		elapsed := time.Now().Sub(startTime)
		if elapsed > maxWait {
			logger.Info(fmt.Sprintf("Agent %s wait timeout for resource %s (elapsed: %v)", agentID, resourceID, elapsed))
			return false, nil
		}

		acquired, err := lm.AcquireLock(ctx, resourceID, agentID, 30*time.Second)
		if err != nil {
			return false, err
		}

		if acquired {
			return true, nil
		}

		time.Sleep(1 * time.Second)
	}
}

func (lm *LockManager) IsLocked(ctx context.Context, resourceID string) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", resourceID)

	exists, err := lm.RedisClient.Exists(ctx, lockKey)
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func (lm *LockManager) GetLockHolder(ctx context.Context, resourceID string) (string, error) {
	lockKey := fmt.Sprintf("lock:%s", resourceID)

	holder, err := lm.RedisClient.Get(ctx, lockKey)
	if err != nil {
		return "", err
	}

	return holder, nil
}

func (lm *LockManager) ForceReleaseLock(ctx context.Context, resourceID string) error {
	lockKey := fmt.Sprintf("lock:%s", resourceID)

	err := lm.RedisClient.Del(ctx, lockKey)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to force release lock: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Force released lock for resource %s", resourceID))
	return nil
}

func (lm *LockManager) ExtendLock(ctx context.Context, resourceID string, additionalTime time.Duration) error {
	lockKey := fmt.Sprintf("lock:%s", resourceID)

	err := lm.RedisClient.Expire(ctx, lockKey, additionalTime)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to extend lock: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Extended lock for resource %s (additional time: %v)", resourceID, additionalTime))
	return nil
}

func (lm *LockManager) GetLockTTL(ctx context.Context, resourceID string) (time.Duration, error) {
	lockKey := fmt.Sprintf("lock:%s", resourceID)

	ttl, err := lm.RedisClient.TTL(ctx, lockKey)
	if err != nil {
		return 0, err
	}

	return ttl, nil
}

type ResourceLock struct {
	ResourceID string
	AgentID    string
	AcquiredAt time.Time
	Timeout    time.Duration
}

func NewResourceLock(resourceID, agentID string, timeout time.Duration) *ResourceLock {
	return &ResourceLock{
		ResourceID: resourceID,
		AgentID:    agentID,
		AcquiredAt: time.Now(),
		Timeout:    timeout,
	}
}

func (rl *ResourceLock) IsExpired() bool {
	elapsed := time.Now().Sub(rl.AcquiredAt)
	return elapsed > rl.Timeout
}

func (rl *ResourceLock) RemainingTime() time.Duration {
	elapsed := time.Now().Sub(rl.AcquiredAt)
	remaining := rl.Timeout - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}
