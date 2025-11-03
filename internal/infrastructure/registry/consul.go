package registry

import (
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/consul/v2"
	"order/internal/config"
	"time"
)

// ConsulRegister consul注册
func ConsulRegister(confInfo *config.ConsulInfo) registry.Registry {
	return consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = confInfo.RegistryAddrs
		options.Timeout = time.Duration(confInfo.Timeout) * time.Second
	})
}
