package main

import (
	"context"
	"github.com/micro/go-micro/v2"
	config2 "github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/util/log"
	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	service2 "github.com/zhanshen02154/order/internal/application/service"
	appconfig "github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/internal/infrastructure/config"
	"github.com/zhanshen02154/order/internal/infrastructure/registry"
	"github.com/zhanshen02154/order/internal/interfaces/handler"
	"github.com/zhanshen02154/order/proto/order"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func main() {
	// 从consul获取配置
	consulConfigSource := config.LoadConsulCOnfig()
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
	err = configInfo.Load(consulConfigSource)
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

	serviceContext, err := infrastructure.NewServiceContext(&confInfo, consulRegistry)
	defer serviceContext.Close()
	if err != nil {
		log.Fatalf("error to load service context: %s", err)
		return
	}

	// 健康检查
	probeServer := infrastructure.NewProbeServer(confInfo.Service.HeathCheckAddr, serviceContext)
	if err = probeServer.Start(); err != nil {
		log.Fatalf("健康检查服务器启动失败")
		return
	}
	if confInfo.Service.Debug {
		runtime.SetBlockProfileRate(1)
		runtime.SetCPUProfileRate(1)
		runtime.SetMutexProfileFraction(1)
		go func() {
			if err = http.ListenAndServe(":6060", nil); err != nil && err != http.ErrServerClosed {
				log.Fatalf("pprof服务器启动失败")
			}
			log.Info("pprof服务器已关闭")
		}()
	}

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
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
			defer cancel()
			log.Info("收到关闭信号，正在停止健康检查服务器...")
			err =  probeServer.Shutdown(shutdownCtx)
			if err != nil {
				return err
			}
			return nil
		}),
	)
	//service.Init()
	orderAppService := service2.NewOrderApplicationService(serviceContext)

	// Register Handler
	err = order.RegisterOrderHandler(service.Server(), &handler.OrderHandler{OrderAppService: orderAppService})
	if err != nil {
		log.Fatal(err)
	}

	if err = service.Run(); err != nil {
		log.Fatal(err)
	}

}