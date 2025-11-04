package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/util/log"
	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	service2 "github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/internal/infrastructure/config"
	"github.com/zhanshen02154/order/internal/infrastructure/persistence"
	gorm2 "github.com/zhanshen02154/order/internal/infrastructure/persistence/gorm"
	"github.com/zhanshen02154/order/internal/infrastructure/registry"
	"github.com/zhanshen02154/order/internal/interfaces/handler"
	"github.com/zhanshen02154/order/proto/order"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	db, err := persistence.InitDB(&confInfo.Database)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	// 健康检查
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	probeServer := infrastructure.NewProbeServer(":8080", db)
	probeServer.Start(ctx)

	txManager := gorm2.NewGormTransactionManager(db)
	orderRepo := gorm2.NewOrderRepository(db)

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

	orderAppService := service2.NewOrderApplicationService(txManager, orderRepo)

	// Register Handler
	err = order.RegisterOrderHandler(service.Server(), &handler.OrderHandler{OrderAppService: orderAppService})
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <- sigChan:
			cancel()
		}
	}()

	// 这里需要单独一个协程来启动服务，
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = service.Run(); err != nil {
			log.Fatal(err)
			cancel()
		}
	}()

	wg.Wait()
	probeServer.Wait()

}


