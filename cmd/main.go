package main

import (
	microzap "github.com/go-micro/plugins/v4/logger/zap"
	"github.com/zhanshen02154/order/internal/bootstrap"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"go-micro.dev/v4/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
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

	zapLogger.With(
		zap.String("service", confInfo.Service.Name),
		zap.String("version", confInfo.Service.Version),
		zap.String("type", "core"),
	)
	loggerMetadataMap["service"] = confInfo.Service.Name
	loggerMetadataMap["version"] = confInfo.Service.Version
	logger.DefaultLogger = logger.DefaultLogger.Fields(loggerMetadataMap)

	if err != nil {
		logger.Error("failed to update logger: ", err)
		return
	}
	//t,io,err := common.NewTracer(ServiceName, "127.0.0.1:6831")
	//if err != nil {
	//	logger.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(t)
	gormLogger := infrastructure.NewGromLogger(zapLogger, confInfo.Database.LogLevel)
	serviceContext, err := infrastructure.NewServiceContext(&confInfo, gormLogger)
	if err != nil {
		logger.Error("error to load service context: ", err)
		return
	}
	defer serviceContext.Close()


	if err := bootstrap.RunService(&confInfo, serviceContext, zapLogger); err != nil {
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