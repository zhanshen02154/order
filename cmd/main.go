package main

import (
	"context"
	microzap "github.com/go-micro/plugins/v4/logger/zap"
	"github.com/zhanshen02154/order/internal/bootstrap"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"go-micro.dev/v4/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"time"
)

func main() {
	zapLogger := zap.New(zapcore.NewCore(getEncoder(), zapcore.AddSync(os.Stdout), zap.InfoLevel),
		zap.WithCaller(true),
		zap.AddCallerSkip(1),
	)
	defer zapLogger.Sync()
	microLogger, err := microzap.NewLogger(microzap.WithLogger(zapLogger))
	if err != nil {
		log.Println(err)
		return
	}
	logger.DefaultLogger = microLogger
	loggerMetadataMap := make(map[string]interface{})

	consulSource, err := config.GetConfig()
	if err != nil {
		logger.Error(err)
		return
	}
	//defer func() {
	//	if consulSource != nil {
	//		if err := consulSource.Close(); err != nil {
	//			logger.Error("failed to close config: ", err)
	//		}
	//	}
	//	return
	//}()

	var confInfo config.SysConfig
	if err := consulSource.Get("order").Scan(&confInfo); err != nil {
		logger.Error(err)
		return
	}

	componentLogger := zapLogger.With(
		zap.String("service", confInfo.Service.Name),
		zap.String("version", confInfo.Service.Version),
	)
	loggerMetadataMap["service"] = confInfo.Service.Name
	loggerMetadataMap["version"] = confInfo.Service.Version
	loggerMetadataMap["type"]    = "core"
	logger.DefaultLogger = logger.DefaultLogger.Fields(loggerMetadataMap)

	if err != nil {
		logger.Error("failed to update logger: ", err)
		return
	}

	// 链路追踪
	traceShutdown := initTracer(confInfo.Service.Name, confInfo.Service.Version, confInfo.Tracer)
	defer traceShutdown()

	//t,io,err := common.NewTracer(ServiceName, "127.0.0.1:6831")
	//if err != nil {
	//	logger.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(t)
	gormLogger := infrastructure.NewGromLogger(componentLogger, confInfo.Database.LogLevel)
	serviceContext, err := infrastructure.NewServiceContext(&confInfo, gormLogger)
	if err != nil {
		logger.Error("error to load service context: ", err)
		return
	}
	defer serviceContext.Close()


	if err := bootstrap.RunService(&confInfo, serviceContext, componentLogger); err != nil {
		logger.Error("failed to start service: ", err)
	}
}

// 获取日志编码器
func getEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(
		zapcore.EncoderConfig{
			MessageKey:          "message",
			LevelKey:            "level",
			TimeKey:             "timestamp",
			NameKey:             "logger",
			CallerKey:           "caller",
			FunctionKey:         zapcore.OmitKey,
			StacktraceKey:       "stacktrace",
			LineEnding:          zapcore.DefaultLineEnding,
			EncodeLevel:         zapcore.LowercaseLevelEncoder,
			EncodeTime:          zapcore.EpochTimeEncoder,
			EncodeDuration:      zapcore.MillisDurationEncoder,
			EncodeCaller:        zapcore.ShortCallerEncoder,
		})
}

// 加载OpenTelemetry链路追踪
func initTracer(serviceName string, version string, conf *config.Tracer) func() {
	ctx := context.Background()
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithTimeout(time.Duration(conf.Client.Timeout) * time.Second),
		otlptracegrpc.WithEndpoint(conf.Client.Endpoint),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled: conf.Client.Retry.Enabled,
			InitialInterval: time.Duration(conf.Client.Retry.InitialInterval) * time.Second,
			MaxInterval: time.Duration(conf.Client.Retry.MaxInterval) * time.Second,
			MaxElapsedTime: time.Duration(conf.Client.Retry.MaxElapsedTime) * time.Second,
		}),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithCompressor("gzip"),
	)

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
		),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithOS(),
	)
	if err != nil {
		logger.Error("failed to create tracer resource: ", err.Error())
	}
	traceExp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		logger.Error("failed to create tracer exporter: ", err.Error())
	}
	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(conf.SampleRate))),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)
	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}