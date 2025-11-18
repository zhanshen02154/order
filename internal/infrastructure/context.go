package infrastructure

import (
	"fmt"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/domain/repository"
	"github.com/zhanshen02154/order/internal/infrastructure/persistence"
	gorm2 "github.com/zhanshen02154/order/internal/infrastructure/persistence/gorm"
	"github.com/zhanshen02154/order/internal/infrastructure/persistence/transaction"
	"github.com/zhanshen02154/order/proto/product"
	"gorm.io/gorm"
	"time"
)

type ServiceContext struct {
	TxManager       transaction.TransactionManager
	LockManager     LockManager
	Conf            *config.SysConfig
	db              *gorm.DB
	OrderRepository repository.IOrderRepository
	ProductClient   product.ProductService
}

func NewServiceContext(conf *config.SysConfig, serviceReg registry.Registry) (*ServiceContext, error) {
	db, err := persistence.InitDB(&conf.Database)
	if err != nil {
		return nil, err
	}

	// 加载ETCD分布式锁
	lockMgr, err := NewEtcdLockManager(&conf.Etcd)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to load lock manager: %v", err))
		return nil, err
	}

	// 初始化商品服务客户端
	grpcClient := grpc.NewClient(
		client.Selector(
			selector.NewSelector(
				selector.Registry(serviceReg),
				selector.SetStrategy(selector.RoundRobin),
			),
		),
		client.Registry(serviceReg),
		client.PoolSize(500),
		client.PoolTTL(5 * time.Minute),
		client.RequestTimeout(5 * time.Second),
		client.DialTimeout(15 * time.Second),
	)
	productClient := product.NewProductService(conf.Consumer.Product.ServiceName, grpcClient)
	return &ServiceContext{
		TxManager:       gorm2.NewGormTransactionManager(db),
		LockManager:     lockMgr,
		Conf:            conf,
		db:              db,
		OrderRepository: gorm2.NewOrderRepository(db),
		ProductClient: productClient,
	}, nil
}

// 关闭所有服务
func (svc *ServiceContext) Close() {
	// 关闭数据库
	if err := svc.closeDB(); err != nil {
		log.Fatalf("close database error: %v", err)
	}
	// 关闭ETCD
	if err := svc.closeEtcd(); err != nil {
		log.Fatalf("close etcd error: %v", err)
	}
}

// 关闭数据库连接
func (svc *ServiceContext) closeDB() error {
	sqlDB, err := svc.db.DB()
	if err != nil {

		return err
	} else {
		log.Info("Preparing to close GORM")
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatalf("Failed to close database instance: %v", err)
		return err
	} else {
		log.Info("GORM数据库连接已关闭")
	}
	return nil
}

// 关闭ETCD
func (svc *ServiceContext) closeEtcd() error {
	err := svc.LockManager.Close()
	if err != nil {
		log.Fatalf("Failed to close etcd lock manager: %v", err)
	} else {
		log.Info("ETCD lock manager closed")
	}
	return err
}

// 检查是否健康
func (svc *ServiceContext) CheckHealth() error {
	sqlDB, err := svc.db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to close database instance: %v", err)
	}
	return nil
}

