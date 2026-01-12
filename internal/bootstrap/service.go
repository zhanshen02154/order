package bootstrap

import (
	"context"
	"github.com/Shopify/sarama"
	kafkabroker "github.com/go-micro/plugins/v4/broker/kafka"
	grpcclient "github.com/go-micro/plugins/v4/client/grpc"
	grpc2 "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
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
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

// RunService 运行服务
func RunService(conf *config.SysConfig, serviceContext *infrastructure.ServiceContext, zapLogger *zap.Logger) error {
	//注册中心
	consulRegistry := infrastructure.ConsulRegister(conf.Consul)

	probeServer := infrastructure.NewProbeServer(conf.Service.HeathCheckAddr, serviceContext)

	var pprofSrv *infrastructure.PprofServer
	if conf.Service.Debug {
		pprofSrv = infrastructure.NewPprofServer(":6060")
	}

	//common.PrometheusBoot(PrometheusPort)

	// New Service
	logWrapper := infrastructure.NewLogWrapper(
		infrastructure.WithZapLogger(zapLogger),
		infrastructure.WithRequestSlowThreshold(conf.Service.RequestSlowThreshold),
		infrastructure.WithSubscribeSlowThreshold(conf.Broker.SubscribeSlowThreshold),
	)
	client := grpcclient.NewClient(
		grpcclient.PoolMaxIdle(100),
	)
	// 为 AsyncProducer 准备 channels，并把它们传给 kafka 插件
	// 使用与 Kafka 配置中相同的缓冲，减少短时写阻塞风险
	successChan := make(chan *sarama.ProducerMessage, conf.Broker.Kafka.ChannelBufferSize)
	errorChan := make(chan *sarama.ProducerError, conf.Broker.Kafka.ChannelBufferSize)
	var eb event.Listener
	broker := infrastructure.NewKafkaBroker(conf.Broker.Kafka, kafkabroker.AsyncProducer(errorChan, successChan))
	deadLetterWrapper := wrapper.NewDeadLetterWrapper(broker)
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

		//添加限流
		micro.WrapHandler(
			ratelimit.NewHandlerWrapper(conf.Service.Qps),
			opentelemetry.NewHandlerWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
			logWrapper.RequestLogWrapper,
		),
		micro.Broker(broker),
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
			if conf.Service.Debug {
				if err := pprofSrv.Close(shutdownCtx); err != nil {
					logger.Error("pprof服务器关闭错误: ", err)
				}
			}

			// 关闭所有系统组件
			serviceContext.Close()
			return nil
		}),
		micro.AfterStop(func() error {
			if eb != nil {
				eb.Close()
			}
			return nil
		}),
		micro.WrapClient(
			opentelemetry.NewClientWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
			wrapper.NewMetaDataWrapper(conf.Service.Name, conf.Service.Version),
		),
		micro.WrapSubscriber(
			opentelemetry.NewSubscriberWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
			logWrapper.SubscribeWrapper(),
			deadLetterWrapper.Wrapper(),
		),
	)
	// 注册应用层服务及事件侦听器（注入 broker 和 producer channels）
	eb = event.NewListener(
		event.WithBroker(broker),
		event.WithClient(service.Client()),
		event.WithLogger(zapLogger),
		event.WithPulishTimeThreshold(conf.Broker.Kafka.Producer.PublishTimeThreshold),
		event.WithProducerChannels(successChan, errorChan),
	)
	event.RegisterPublisher(conf.Broker, eb)
	eb.Start()
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
