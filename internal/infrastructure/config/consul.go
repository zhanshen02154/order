package config

import (
	"fmt"
	"github.com/micro/go-micro/v2/config/encoder/yaml"
	"github.com/micro/go-micro/v2/config/source"
	"github.com/micro/go-plugins/config/source/consul/v2"
	"github.com/zhanshen02154/order/pkg/env"
)

func LoadConsulCOnfig() source.Source {
	// 从consul获取配置
	consulHost := env.GetEnv("CONSUL_HOST", "192.168.83.131")
	consulPort := env.GetEnv("CONSUL_PORT", "8500")
	consulPrefix := env.GetEnv("CONSUL_PREFIX", "/micro/")
	consulSource := consul.NewSource(
		// Set configuration address
		consul.WithAddress(fmt.Sprintf("%s:%s", consulHost, consulPort)),
		//前缀 默认：/micro/product
		consul.WithPrefix(consulPrefix),
		//consul.StripPrefix(true),
		source.WithEncoder(yaml.NewEncoder()),
		consul.StripPrefix(true),
	)
	return consulSource
}
