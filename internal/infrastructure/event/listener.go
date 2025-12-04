package event

import (
	"context"
	"github.com/zhanshen02154/order/internal/config"
	"go-micro.dev/v4/client"
)

// 事件总线
type Listener interface {
	Publish(ctx context.Context, topic string, event interface{}, key interface{}, opts ...client.PublishOption) error
	Register(topic string) bool
	UnRegister(topic string) bool
	Close()
}

// 注册发布事件
func RegisterPublisher(conf *config.Broker, eb Listener) {
	if len(conf.Publisher) > 0 {
		for i := range conf.Publisher {
			eb.Register(conf.Publisher[i])
		}
	}
}