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
	etcdConf := &config.Etcd{
		Hosts:            []string{"http://127.0.0.1:2379"},
		DialTimeout:      30,
		Username:         "order",
		Password:         "",
		AutoSyncInterval: 500,
		Prefix:           "/micro/order/",
	}
	lockManager, _ = NewEtcdLockManager(etcdConf)
}

func TestLock(t *testing.T) {
	setup()
	lockkey := "testKey"
	ctx := context.Background()
	lock, err := lockManager.NewLock(ctx, lockkey)
	if err != nil {
		logger.Info(err)
		return
	}
	flag, err := lock.TryLock(ctx)
	unlockFlag, unlockErr := lock.UnLock(ctx)
	assert.Equal(t, true, flag)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, unlockFlag)
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
