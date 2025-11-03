package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/util/log"
	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	"net/http"
	"net/url"
	service2 "github.com/zhanshen02154/order/internal/application/service"
	configstruct "github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure/config"
	gormrepo "github.com/zhanshen02154/order/internal/infrastructure/persistence/gorm"
	"github.com/zhanshen02154/order/internal/infrastructure/registry"
	"github.com/zhanshen02154/order/internal/interfaces/handler"
	order "github.com/zhanshen02154/order/proto/order"
	"time"
)

func main() {
	confInfo, err := config.LoadSystemConfig()
	if err != nil {
		panic(err)
	}

	//注册中心
	consulRegistry := registry.ConsulRegister(&confInfo.Consul)

	//t,io,err := common.NewTracer(ServiceName, "127.0.0.1:6831")
	//if err != nil {
	//	log.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(t)

	db, err := initDB(&confInfo.Database)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	defer func() {
		if db != nil { // 关键检查
			db.Close()
		}
	}()

	tableInit := gormrepo.NewOrderRepository(db)
	//tableInit.InitTable()

	//common.PrometheusBoot(PrometheusPort)

	// New Service
	service := micro.NewService(
		micro.Name(confInfo.Service.Name),
		micro.Version(confInfo.Service.Version),
		micro.Address(confInfo.Service.Listen),
		micro.Registry(consulRegistry),
		micro.RegisterTTL(time.Duration(confInfo.Consul.RegisterTtl)*time.Second),
		micro.RegisterInterval(time.Duration(confInfo.Consul.RegisterInterval)*time.Second),
		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		//添加限流
		micro.WrapHandler(ratelimit.NewHandlerWrapper(confInfo.Service.Qps)),
		//添加监控
		//micro.WrapHandler(prometheus.NewHandlerWrapper()),
	)

	// Initialise service
	service.Init()

	// 健康检查
	go func() {
		// livenessProbe
		http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
		})

		// readinessProbe
		http.HandleFunc("/ready", func(writer http.ResponseWriter, request *http.Request) {
			if err := db.DB().Ping(); err != nil {
				writer.WriteHeader(http.StatusServiceUnavailable)
				writer.Write([]byte("Not Ready"))
			} else {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("Ok"))
			}
		})
		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("check status http api error: %v", err)
		} else {
			log.Info("listen http server on: 8080")
		}
	}()

	orderAppService := service2.NewOrderApplicationService(tableInit)

	// Register Handler
	err = order.RegisterOrderHandler(service.Server(), &handler.OrderHandler{OrderAppService: orderAppService})
	if err != nil {
		log.Fatal(err)
	}

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

// 初始化数据库
func initDB(confInfo *configstruct.MySqlConfig) (*gorm.DB, error) {
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
	sqlDB.SetMaxOpenConns(confInfo.MaxOpenConns)
	sqlDB.SetMaxIdleConns(confInfo.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(confInfo.ConnMaxLifeTime) * time.Second)

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接验证失败: %w", err)
	}

	log.Info("数据库连接成功")
	return db, nil
}
