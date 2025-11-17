package tests

import (
	"context"
	"github.com/bmizerany/assert"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"testing"
)

var lockManager infrastructure.LockManager

func setup() {
	etcdConf := &config.Etcd{
		Hosts:            []string{"http://192.168.83.131:2379"},
		DialTimeout:      30,
		Username:         "order",
		Password:         "347834dh",
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
	lock := lockManager.NewLock(ctx, lockkey)
	flag, err := lock.Lock(ctx, 30)
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
		log.Fatal(err)
	}else {
		log.Info("lock manager closed")
	}
}