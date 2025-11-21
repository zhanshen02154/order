package main

import (
	"context"
	grpc2 "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	appservice "github.com/zhanshen02154/order/internal/application/service"
	config2 "github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	config3 "github.com/zhanshen02154/order/internal/infrastructure/config"
	"github.com/zhanshen02154/order/internal/infrastructure/registry"
	"github.com/zhanshen02154/order/internal/interfaces/handler"
	"github.com/zhanshen02154/order/pkg/codec"
	"github.com/zhanshen02154/order/proto/order"
	"go-micro.dev/v4"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"net/http"
	"runtime"
	"time"
)

func main() {
	consulSource := config3.LoadConsulCOnfig()
	configInfo, err := config.NewConfig()
	defer func() {
		err = configInfo.Close()
		if err != nil {
			logger.Fatal(err)
		}
	}()
	if err != nil {
		logger.Fatal(err)
		return
	}
	err = configInfo.Load(consulSource)
	if err != nil {
		logger.Fatal(err)
		return
	}

	var confInfo config2.SysConfig
	if err := configInfo.Get("order").Scan(&confInfo); err != nil {
		logger.Fatal(err)
	}

	//注册中心
	consulRegistry := registry.ConsulRegister(&confInfo.Consul)

	//t,io,err := common.NewTracer(ServiceName, "127.0.0.1:6831")
	//if err != nil {
	//	logger.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(t)

	serviceContext, err := infrastructure.NewServiceContext(&confInfo)
	defer serviceContext.Close()
	if err != nil {
		logger.Fatalf("error to load service context: %s", err)
		return
	}

	// 健康检查
	probeServer := infrastructure.NewProbeServer(confInfo.Service.HeathCheckAddr, serviceContext)
	if err = probeServer.Start(); err != nil {
		logger.Fatalf("健康检查服务器启动失败")
		return
	}
	if confInfo.Service.Debug {
		runtime.SetBlockProfileRate(1)
		runtime.SetCPUProfileRate(1)
		runtime.SetMutexProfileFraction(1)
		go func() {
			if err := http.ListenAndServe(":6060", nil); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("pprof服务器启动失败: %s", err)
				return
			}else {
				logger.Info("pprof启动成功")
			}
			logger.Info("pprof服务器已关闭")
			return
		}()
	}
	//tableInit.InitTable()

	//common.PrometheusBoot(PrometheusPort)

	// New Service
	service := micro.NewService(
		micro.Server(grpc2.NewServer(
			server.Name(confInfo.Service.Name),
			server.Version(confInfo.Service.Version),
			server.Address(confInfo.Service.Listen),
			server.Transport(grpc.NewTransport()),
			server.Registry(consulRegistry),
			server.RegisterTTL(time.Duration(confInfo.Consul.RegisterTtl)*time.Second),
			server.RegisterInterval(time.Duration(confInfo.Consul.RegisterInterval)*time.Second),
			//server.WrapHandler(dtm.NewHandlerWrapper),
			grpc2.Codec("application/grpc+dtm_raw", codec.NewDtmCodec()),
			)),
		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		//添加限流
		micro.WrapHandler(ratelimit.NewHandlerWrapper(confInfo.Service.Qps)),

		//添加监控
		//micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.BeforeStop(func() error {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			logger.Info("收到关闭信号，正在停止健康检查服务器...")
			err = probeServer.Shutdown(shutdownCtx)
			if err != nil {
				return err
			}
			return nil
		}),
	)
	//service.Init()
	orderAppService := appservice.NewOrderApplicationService(serviceContext)

	// Register Handler
	err = order.RegisterOrderHandler(service.Server(), &handler.OrderHandler{OrderAppService: orderAppService})
	if err != nil {
		logger.Fatal(err)
	}

	if err = service.Run(); err != nil {
		logger.Fatal(err)
	}
}