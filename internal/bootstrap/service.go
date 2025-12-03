package bootstrap

import (
	"context"
	grpc2 "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	appservice "github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/internal/infrastructure/broker/kafka"
	event2 "github.com/zhanshen02154/order/internal/infrastructure/event"
	"github.com/zhanshen02154/order/internal/infrastructure/event/wrapper"
	"github.com/zhanshen02154/order/internal/infrastructure/registry"
	localserver "github.com/zhanshen02154/order/internal/infrastructure/server"
	"github.com/zhanshen02154/order/internal/interfaces/event"
	"github.com/zhanshen02154/order/internal/interfaces/handler"
	"github.com/zhanshen02154/order/pkg/codec"
	"github.com/zhanshen02154/order/proto/order"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"time"
)

func RunService(conf *config.SysConfig, serviceContext *infrastructure.ServiceContext) error {
	//注册中心
	consulRegistry := registry.ConsulRegister(conf.Consul)

	probeServer := localserver.NewProbeServer(conf.Service.HeathCheckAddr, serviceContext)

	var pprofSrv *localserver.PprofServer
	if conf.Service.Debug {
		pprofSrv = localserver.NewPprofServer(":6060")
	}
	//tableInit.InitTable()

	//common.PrometheusBoot(PrometheusPort)

	// New Service
	broker := kafka.NewKafkaBroker(conf.Broker.Kafka)
	service := micro.NewService(
		micro.Server(grpc2.NewServer(
			server.Name(conf.Service.Name),
			server.Version(conf.Service.Version),
			server.Address(conf.Service.Listen),
			server.Transport(grpc.NewTransport()),
			server.Registry(consulRegistry),
			server.RegisterTTL(time.Duration(conf.Consul.RegisterTtl)*time.Second),
			server.RegisterInterval(time.Duration(conf.Consul.RegisterInterval)*time.Second),
			grpc2.Codec("application/grpc+dtm_raw", codec.NewDtmCodec()),
		)),
		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		//添加限流
		micro.WrapHandler(ratelimit.NewHandlerWrapper(conf.Service.Qps)),
		micro.Broker(broker),

		//添加监控
		//micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.AfterStart(func() error {
			pprofSrv.Start()
			//if err := broker.Connect(); err != nil {
			//	return err
			//}
			if err := probeServer.Start(); err != nil {
				logger.Error("健康检查服务器启动失败")
			}
			return nil
		}),
		micro.BeforeStop(func() error {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			logger.Info("收到关闭信号，正在停止健康检查服务器...")
			err := probeServer.Shutdown(shutdownCtx)
			if err != nil {
				return err
			}
			if conf.Service.Debug {
				if err := pprofSrv.Close(shutdownCtx); err != nil {
					logger.Error("pprof服务器关闭错误: ", err)
				}
			}

			return nil
		}),
		micro.WrapClient(wrapper.NewMetaDataWrapper(conf.Service.Name, conf.Service.Version)),
	)

	// 注册应用层服务及事件侦听器
	eb := event2.NewListener(service.Client())
	registerPublisher(conf.Broker, eb)
	defer eb.Close()
	orderAppService := appservice.NewOrderApplicationService(serviceContext, eb)

	orderEventHandler := event.NewHandler(orderAppService)
	// 注册订阅事件
	if len(conf.Broker.Subscriber) > 0 {
		for i := range conf.Broker.Subscriber {
			if err := micro.RegisterSubscriber(conf.Broker.Subscriber[i], service.Server(), orderEventHandler); err != nil {
				logger.Error("failed to register subsriber: ", err)
				continue
			}
		}
	}

	// Register Handler
	err := order.RegisterOrderHandler(service.Server(), &handler.OrderHandler{OrderAppService: orderAppService})
	if err != nil {
		logger.Error(err)
	}

	if err = service.Run(); err != nil {
		logger.Error(err)
	}

	return nil
}

// 注册发布事件
func registerPublisher(conf *config.Broker, eb event2.Bus) {
	if len(conf.Publisher) > 0 {
		for i := range conf.Publisher {
			eb.Register(conf.Publisher[i])
		}
	}
}