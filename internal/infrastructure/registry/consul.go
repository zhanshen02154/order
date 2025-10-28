package registry

import (
	"git.imooc.com/zhanshen1614/order/internal/config"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/consul/v2"
	"time"
)

// ConsulRegister consul注册
func ConsulRegister(confInfo *config.ConsulInfo) registry.Registry {
	return consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = confInfo.RegistryAddrs
		options.Timeout = time.Duration(confInfo.Timeout) * time.Second
	})
}
