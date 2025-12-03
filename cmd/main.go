package main

import (
	"github.com/zhanshen02154/order/internal/bootstrap"
	"github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"go-micro.dev/v4/logger"
)

func main() {
	consulSource, err := config.GetConfig()
	if err != nil {
		logger.Error(err)
		return
	}
	defer func() {
		if consulSource != nil {
			if err := consulSource; err != nil {
				logger.Error("failed to close config: ", err)
			}
		}
		return
	}()

	var confInfo config.SysConfig
	if err := consulSource.Get("order").Scan(&confInfo); err != nil {
		logger.Error(err)
	}
	//t,io,err := common.NewTracer(ServiceName, "127.0.0.1:6831")
	//if err != nil {
	//	logger.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(t)

	serviceContext, err := infrastructure.NewServiceContext(&confInfo)
	if err != nil {
		logger.Error("error to load service context: ", err)
		return
	}
	defer serviceContext.Close()


	if err := bootstrap.RunService(&confInfo, serviceContext); err != nil {
		logger.Error("failed to start service: ", err)
	}
}
