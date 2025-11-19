package tests

import (
	"context"
	"github.com/bmizerany/assert"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"go-micro.dev/v4/logger"
	"testing"
)

var lockManager infrastructure.LockManager

func setup() {
	etcdConf := &config.Etcd{
		Hosts:            []string{"http://127.0.0.1:2379"},
		DialTimeout:      30,
		Username:         "order",
		Password:         "",
		AutoSyncInterval: 5,
		Prefix:           "/micro/order/",
	}
	lockManager, _ = infrastructure.NewEtcdLockManager(etcdConf)
}

func TestLock(t *testing.T) {
	setup()
	defer teardown()
	lockkey := "testKey"
	ctx := context.Background()
	lock, err := lockManager.NewLock(ctx, lockkey, 30)
	if err != nil {
		logger.Info(err)
		return
	}
	flag, err := lock.Lock(ctx)
	defer func() {
		if lock != nil {
			lock.UnLock(ctx)
		}
	}()
	assert.Equal(t, true, flag)
	assert.Equal(t, nil, err)
}

func teardown() {
	err := lockManager.Close()
	if err != nil {
		logger.Fatal(err)
	} else {
		logger.Info("lock manager closed")
	}
}
