package infrastructure

import (
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/domain/repository"
	gorm2 "github.com/zhanshen02154/order/internal/infrastructure/persistence/gorm"
	"github.com/zhanshen02154/order/internal/infrastructure/persistence/transaction"
	"github.com/zhanshen02154/order/internal/infrastructure/persistence/transaction/dtm"
	"go-micro.dev/v4/logger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type ServiceContext struct {
	TxManager       transaction.TransactionManager
	LockManager     LockManager
	Conf            *config.SysConfig
	db              *gorm.DB
	OrderRepository repository.IOrderRepository
	Dtm             *dtm.Server
}

// NewServiceContext 初始化服务上下文
func NewServiceContext(conf *config.SysConfig, zapLogger gormlogger.Interface) (*ServiceContext, error) {
	var err error
	db, err := InitDB(conf.Database, zapLogger)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if sqlDb, err := db.DB(); err == nil {
				sqlDb.Close()
			}
			logger.Error("failed to load service context: " + err.Error())
		}
	}()
	// 加载Redis分布式锁
	lockMgr, err := NewRedisLockManager(conf.Redis)
	if err != nil {
		return nil, err
	}
	return &ServiceContext{
		TxManager:       gorm2.NewGormTransactionManager(db),
		LockManager:     lockMgr,
		Conf:            conf,
		db:              db,
		OrderRepository: gorm2.NewOrderRepository(db),
		Dtm:             dtm.NewServer(conf.Transaction.Host),
	}, nil
}

// Close 关闭所有服务
func (svc *ServiceContext) Close() {
	// 关闭数据库
	if err := svc.closeDB(); err != nil {
		logger.Error("close database error: " + err.Error())
	}
	// 关闭ETCD
	if err := svc.closeLock(); err != nil {
		logger.Error("close redis lock manager error: " + err.Error())
	}
}

// 关闭数据库连接
func (svc *ServiceContext) closeDB() error {
	sqlDB, err := svc.db.DB()
	if err != nil {

		return err
	} else {
		logger.Info("Preparing to close GORM")
	}
	if err := sqlDB.Close(); err != nil {
		logger.Error("Failed to close database instance: " + err.Error())
		return err
	} else {
		logger.Info("Database instance closed")
	}
	return nil
}

// 关闭分布式锁
func (svc *ServiceContext) closeLock() error {
	err := svc.LockManager.Close()
	if err != nil {
		logger.Error("Failed to close redis lock manager: " + err.Error())
	} else {
		logger.Info("Redis lock manager closed")
	}
	return err
}

// CheckHealth 健康检查
func (svc *ServiceContext) CheckHealth() error {
	sqlDB, err := svc.db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Database instance unhealthy: " + err.Error())
	}
	if err := svc.LockManager.CheckHealth(); err != nil {
		logger.Error("Redis lock manager unhealthy: " + err.Error())
		return err
	}
	return nil
}
