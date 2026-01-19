package bootstrap

import (
	"context"
	"github.com/Shopify/sarama"
	kafkabroker "github.com/go-micro/plugins/v4/broker/kafka"
	grpcclient "github.com/go-micro/plugins/v4/client/grpc"
	grpc2 "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	appservice "github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/internal/infrastructure/event"
	"github.com/zhanshen02154/order/internal/infrastructure/event/monitor"
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

	monitorSvr := infrastructure.NewMonitorServer(":6060")

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
			prometheus.NewHandlerWrapper(prometheus.ServiceName(conf.Service.Name), prometheus.ServiceVersion(conf.Service.Version)),
			logWrapper.RequestLogWrapper,
		),
		micro.Broker(broker),
		micro.AfterStart(func() error {
			if monitorSvr != nil {
				monitorSvr.Start()
			}
			if err := probeServer.Start(); err != nil {
				logger.Error("failed to start probe server" + err.Error())
			}
			eb.Start()
			return nil
		}),
		micro.BeforeStop(func() error {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			logger.Info("Stopping monitor servers...")
			if err := monitorSvr.Close(shutdownCtx); err != nil {
				logger.Error("failed to close monitor servers" + err.Error())
			} else {
				logger.Info("Stopping monitor servers successfully")
			}
			if err := probeServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("Failed to close probe server" + err.Error())
			} else {
				logger.Info("Successfully closed monitor servers")
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
			monitor.NewClientWrapper(monitor.WithName(conf.Service.Name), monitor.WithVersion(conf.Service.Version)),
			wrapper.NewMetaDataWrapper(conf.Service.Name, conf.Service.Version),
		),
		micro.WrapSubscriber(
			opentelemetry.NewSubscriberWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
			prometheus.NewSubscriberWrapper(prometheus.ServiceName(conf.Service.Name), prometheus.ServiceVersion(conf.Service.Version)),
			logWrapper.SubscribeWrapper(),
			deadLetterWrapper.Wrapper(),
		),
	)
	// 注册应用层服务及事件侦听器
	eb = event.NewListener(
		event.WithProducerChannels(successChan, errorChan),
		event.WithServiceName(conf.Service.Name),
		event.WithServiceVersion(conf.Service.Version),
		event.WrapPublishCallback(
			event.NewTracerWrapper(event.WithTracerProvider(otel.GetTracerProvider())),
			event.NewPublicCallbackLogWrapper(
				event.WithLogger(zapLogger),
				event.WithTimeThreshold(conf.Broker.PublishTimeThreshold),
			),
			event.NewDeadletterWrapper(event.WithBroker(broker), event.WithTracer(otel.GetTracerProvider()), event.WithServiceInfo(conf.Service)),
		),
	)
	event.RegisterPublisher(conf.Broker, eb, service.Client())
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
