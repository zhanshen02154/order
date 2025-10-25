package main

import (
	"fmt"
	"git.imooc.com/zhanshen1614/common"
	"git.imooc.com/zhanshen1614/order/domain/repository"
	service2 "git.imooc.com/zhanshen1614/order/domain/service"
	"git.imooc.com/zhanshen1614/order/handler"
	order "git.imooc.com/zhanshen1614/order/proto/order"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/micro/go-plugins/registry/consul/v2"
	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	"net/url"
	"time"
)

var (
	QPS         = 1000
	ServiceName = "go.micro.service.order"
	PrometheusPort = 9092
)

func main() {
	consulConfig, err := common.GetConsulConfig("127.0.0.1", 8500, "/micro/config")
	if err != nil {
		log.Error(err)
	}
	//注册中心
	consulRegistry := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{
			"127.0.0.1:8500",
		}
		options.Timeout = 3 * time.Minute
	})

	//t,io,err := common.NewTracer(ServiceName, "127.0.0.1:6831")
	//if err != nil {
	//	log.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(t)

	mysqConfig, err := common.GetMySqlFromConsul(consulConfig, "mysql")
	if err != nil {
		log.Error(err)
	}
	db, err := initDB(mysqConfig)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	defer func() {
		if db != nil { // 关键检查
			db.Close()
		}
	}()

	tableInit := repository.NewOrderRepository(db)
	//tableInit.InitTable()

	//common.PrometheusBoot(PrometheusPort)

	// New Service
	service := micro.NewService(
		micro.Name(ServiceName),
		micro.Version("latest"),
		micro.Address(":9085"),
		micro.Registry(consulRegistry),
		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		//添加限流
		micro.WrapHandler(ratelimit.NewHandlerWrapper(QPS)),
		//添加监控
		//micro.WrapHandler(prometheus.NewHandlerWrapper()),
	)

	// Initialise service
	service.Init()

	orderDataService := service2.NewOrderDataService(tableInit)

	// Register Handler
	order.RegisterOrderHandler(service.Server(), &handler.Order{OrderDataService: orderDataService})

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

// 初始化数据库
func initDB(confInfo *common.MySqlConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		confInfo.User,
		confInfo.Password,
		confInfo.Host,
		confInfo.Port,
		confInfo.Database,
		confInfo.Charset,
		url.QueryEscape(confInfo.Loc),
	)
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	sqlDB := db.DB()
	if sqlDB == nil {
		return nil, fmt.Errorf("获取SQL DB失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(1000)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接验证失败: %w", err)
	}

	log.Info("数据库连接成功")
	return db, nil
}