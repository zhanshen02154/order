package main

import (
	"github.com/zhanshen02154/order/internal/bootstrap"
	config2 "github.com/zhanshen02154/order/internal/config"
	"github.com/zhanshen02154/order/internal/infrastructure"
	config3 "github.com/zhanshen02154/order/internal/infrastructure/config"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/logger"
)

func main() {
	consulSource := config3.LoadConsulCOnfig()
	configInfo, err := config.NewConfig()
	if err != nil {
		logger.Error(err)
		return
	}
	defer func() {
		err = configInfo.Close()
		if err != nil {
			logger.Error(err)
		}
	}()
	err = configInfo.Load(consulSource)
	if err != nil {
		logger.Error(err)
		return
	}

	var confInfo config2.SysConfig
	if err := configInfo.Get("order").Scan(&confInfo); err != nil {
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
