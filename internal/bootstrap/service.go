package bootstrap

import (
	"context"
	grpcclient "github.com/go-micro/plugins/v4/client/grpc"
	grpc2 "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	appservice "github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/internal/infrastructure/event"
	"github.com/zhanshen02154/order/internal/infrastructure/event/wrapper"
	"github.com/zhanshen02154/order/internal/interfaces/handler"
	"github.com/zhanshen02154/order/internal/interfaces/subscriber"
	"github.com/zhanshen02154/order/pkg/codec"
	"github.com/zhanshen02154/order/proto/order"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"go.uber.org/zap"
	"time"
)

func RunService(conf *config.SysConfig, serviceContext *infrastructure.ServiceContext, zapLogger *zap.Logger) error {
	//注册中心
	consulRegistry := infrastructure.ConsulRegister(conf.Consul)

	probeServer := infrastructure.NewProbeServer(conf.Service.HeathCheckAddr, serviceContext)

	var pprofSrv *infrastructure.PprofServer
	if conf.Service.Debug {
		pprofSrv = infrastructure.NewPprofServer(":6060")
	}

	//tableInit.InitTable()

	//common.PrometheusBoot(PrometheusPort)

	// New Service
	logWrapper := infrastructure.NewLogWrapper(zapLogger)
	client := grpcclient.NewClient(
		grpcclient.PoolMaxIdle(100),
		)
	broker := infrastructure.NewKafkaBroker(conf.Broker.Kafka)
	dealLetterWrapper := wrapper.NewDeadLetterWrapper(broker)
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
		micro.Client(client),

		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		//添加限流
		micro.WrapHandler(
			logWrapper.RequestLogWrapper,
			ratelimit.NewHandlerWrapper(conf.Service.Qps),
			),
		micro.Broker(broker),
		//添加监控
		//micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.AfterStart(func() error {
			if pprofSrv != nil {
				pprofSrv.Start()
			}
			if err := broker.Connect(); err != nil {
				return err
			}
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
		micro.WrapClient(wrapper.NewMetaDataWrapper(conf.Service.Name, conf.Service.Version, zapLogger)),
		micro.WrapSubscriber(
			dealLetterWrapper.Wrapper(),
			logWrapper.SubscribeWrapper(),
		),
	)

	// 注册应用层服务及事件侦听器
	eb := event.NewListener(service.Client())
	event.RegisterPublisher(conf.Broker, eb)
	defer eb.Close()
	orderAppService := appservice.NewOrderApplicationService(serviceContext, eb)

	productEventHandler := subscriber.NewProductEventHandler(orderAppService)
	// 注册订阅事件
	productEventHandler.RegisterSubscriber(service.Server())

	// Register Handler
	err := order.RegisterOrderHandler(service.Server(), &handler.OrderHandler{OrderAppService: orderAppService})
	if err != nil {
		logger.Error(err)
		return err
	}

	if err = service.Run(); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}