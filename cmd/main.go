package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client/selector"
	config2 "github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/encoder/yaml"
	"github.com/micro/go-micro/v2/config/source"
	"github.com/micro/go-micro/v2/transport/grpc"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/micro/go-plugins/config/source/consul/v2"
	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	service2 "github.com/zhanshen02154/order/internal/application/service"
	appconfig "github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/internal/infrastructure/persistence"
	gorm2 "github.com/zhanshen02154/order/internal/infrastructure/persistence/gorm"
	"github.com/zhanshen02154/order/internal/infrastructure/registry"
	"github.com/zhanshen02154/order/internal/interfaces/handler"
	"github.com/zhanshen02154/order/pkg/env"
	"github.com/zhanshen02154/order/proto/order"
	"github.com/zhanshen02154/order/proto/product"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	// 从consul获取配置
	consulHost := env.GetEnv("CONSUL_HOST", "192.168.83.131")
	consulPort := env.GetEnv("CONSUL_PORT", "8500")
	consulPrefix := env.GetEnv("CONSUL_PREFIX", "/micro/")
	consulSource := consul.NewSource(
		// Set configuration address
		consul.WithAddress(fmt.Sprintf("%s:%s", consulHost, consulPort)),
		//前缀 默认：/micro/product
		consul.WithPrefix(consulPrefix),
		//consul.StripPrefix(true),
		source.WithEncoder(yaml.NewEncoder()),
		consul.StripPrefix(true),
	)
	configInfo, err := config2.NewConfig()
	defer func() {
		err = configInfo.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	if err != nil {
		log.Fatal(err)
		return
	}
	err = configInfo.Load(consulSource)
	if err != nil {
		log.Fatal(err)
		return
	}

	var confInfo appconfig.SysConfig
	if err = configInfo.Get("order").Scan(&confInfo); err != nil {
		log.Fatal(err)
		return
	}

	//注册中心
	consulRegistry := registry.ConsulRegister(&confInfo.Consul)

	//t,io,err := common.NewTracer(ServiceName, "127.0.0.1:6831")
	//if err != nil {
	//	log.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(t)

	db, err := persistence.InitDB(&confInfo.Database)
	if err != nil {
		log.Fatalf("failed to load gorm framework", err)
		return
	}

	// 健康检查
	probeServer := infrastructure.NewProbeServer(confInfo.Service.HeathCheckAddr, db)
	if err = probeServer.Start(); err != nil {
		log.Fatalf("健康检查服务器启动失败")
		return
	}
	if confInfo.Service.Debug {
		go func() {
			if err = http.ListenAndServe(":6060", nil); err != nil && err != http.ErrServerClosed {
				log.Fatalf("pprof服务器启动失败")
			}
			log.Info("pprof服务器已关闭")
		}()
	}

	txManager := gorm2.NewGormTransactionManager(db)
	orderRepo := gorm2.NewOrderRepository(db)
	// 初始化商品服务客户端
	productService := micro.NewService(
		micro.Name(confInfo.Consumer.Product.ClientName),
		micro.Registry(consulRegistry),
		micro.Selector(selector.NewSelector(selector.Registry(consulRegistry), selector.SetStrategy(selector.RoundRobin))),
		micro.Transport(grpc.NewTransport()),
	)
	productService.Init()
	productClient := product.NewProductService(confInfo.Consumer.Product.ServiceName, productService.Client())

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
		micro.BeforeStop(func() error {
			log.Info("收到关闭信号，正在停止健康检查服务器...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
			defer cancel()
			err =  probeServer.Shutdown(shutdownCtx)
			if err != nil {
				return err
			}
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			if err = sqlDB.Ping(); err == nil {
				err1 := sqlDB.Close()
				if err1 != nil {
					log.Infof("关闭GORM连接失败： %v", err1)
					return err1
				}
			}else {
				log.Info("数据库已关闭")
			}
			return nil
		}),
	)

	// Initialise service
	service.Init()
	orderAppService := service2.NewOrderApplicationService(txManager, orderRepo, productClient)

	// Register Handler
	err = order.RegisterOrderHandler(service.Server(), &handler.OrderHandler{OrderAppService: orderAppService})
	if err != nil {
		log.Fatal(err)
	}

	if err = service.Run(); err != nil {
		log.Fatal(err)
	}

}