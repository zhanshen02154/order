package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/zhanshen02154/order/internal/config"
	"log"
	"time"
)

// 分布式锁接口
type DistributedLock interface {
	Lock(ctx context.Context, ttl int) (bool, error)
	UnLock(ctx context.Context) (bool, error)
	GetKey(ctx context.Context) string
}

// ETCD锁
type EtcdLock struct {
	ecli *clientv3.Client
	session *concurrency.Session
	mutex *concurrency.Mutex
	prefix string
	isLocked bool
}

// 获取键名
func (l *EtcdLock) GetKey(ctx context.Context) string {
	return l.mutex.Key()
}

// 加锁
func (l *EtcdLock) Lock(ctx context.Context, ttl int) (bool, error) {
	if l.isLocked {
		return false, errors.New(fmt.Sprintf("key: %s was locked", l.prefix))
	}
	session, err := concurrency.NewSession(l.ecli, concurrency.WithTTL(ttl), concurrency.WithContext(ctx))
	if err != nil {
		return false, err
	}
	l.session = session
	l.mutex = concurrency.NewMutex(l.session, l.prefix)
	if err = l.mutex.Lock(ctx); err != nil {
		err = l.session.Close()
		if err != nil {
			return false, errors.New(fmt.Sprintf("prefix key: %s session close failed: %s", l.prefix, err))
		}
		return false, err
	}
	l.isLocked = true
	return true, nil
}

// 解锁
func (l *EtcdLock) UnLock(ctx context.Context) (bool, error) {
	defer func() {
		err := l.session.Close()
		if err != nil {
			log.Fatalf(fmt.Sprintf("prefix key: %s session close failed: %s", l.prefix, err))
		}
	}()
	if err := l.mutex.Unlock(ctx); err != nil {
		return false, err
	}
	l.isLocked = false
	return true, nil
}

// 分布式锁管理器
type LockManager interface {
	NewLock(ctx context.Context, key string) DistributedLock
	Close() error
}

// ETCD分布式锁
type EtcdLockManager struct {
	ecli *clientv3.Client
	prefix string
}

// 关闭客户端
func (elm *EtcdLockManager) Close() error {
	return elm.ecli.Close()
}

// 创建锁
func (elm *EtcdLockManager) NewLock(ctx context.Context, key string) DistributedLock {
	return &EtcdLock{
		ecli:     elm.ecli,
		prefix:   fmt.Sprintf("%slock/%s/", elm.prefix, key),
		isLocked: false,
	}
}

// 创建分布式锁
func NewEtcdLockManager(conf *config.Etcd) (LockManager, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:            conf.Hosts,
		AutoSyncInterval:     time.Duration(conf.AutoSyncInterval) * time.Second,
		DialTimeout:          time.Duration(conf.DialTimeout) * time.Second,
		Username:             conf.Username,
		Password:             conf.Password,
	})
	if err != nil {
		return nil, err
	}
	log.Println("ETCD was stared")
	return &EtcdLockManager{ecli: client, prefix: conf.Prefix}, nil
}

