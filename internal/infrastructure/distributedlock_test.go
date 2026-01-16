package infrastructure

import (
	"context"
	"github.com/bmizerany/assert"
	"github.com/zhanshen02154/order/internal/config"
	"go-micro.dev/v4/logger"
	"testing"
)

var lockManager LockManager

func setup() {
	conf := &config.Redis{
		Addr:           "127.0.0.1:6379",
		Password:       "",
		Database:       1,
		PoolSize:       2,
		DialTimeout:    10,
		ReadTimeout:    10,
		WriteTimeout:   5,
		MinIdleConns:   1,
		Prefix:         "order",
		LockTries:      3,
		LockRetryDelay: 500,
	}
	lockManager, _ = NewRedisLockManager(conf)
	if lockManager == nil {
		return
	}
}

func TestLock(t *testing.T) {
	setup()
	lockkey := "testKey"
	ctx := context.Background()
	lock := lockManager.NewLock(lockkey, 10)
	err := lock.TryLock(ctx)
	unlockErr := lock.UnLock(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, nil, unlockErr)
}

func teardown() {
	err := lockManager.Close()
	if err != nil {
		logger.Error(err)
	} else {
		logger.Info("lock manager closed")
	}
}
