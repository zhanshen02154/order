package event

import (
	"context"
	"encoding/json"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/proto/order"
	"go-micro.dev/v4/logger"
)

// 订单事件处理器接口
type OrderEventHandler interface {
	ReceiveData(ctx context.Context, msg *order.OrderPaid) error
}

// 订单事件处理器实现类
type OrderEventHandlerImpl struct {
	orderAppService service.IOrderApplicationService
}

func (h *OrderEventHandlerImpl) ReceiveData(ctx context.Context, msg *order.OrderPaid) error {
	b, err := json.Marshal(msg)
	if err != nil {
		logger.Error(err)
	}else {
		logger.Info("received data: ", string(b))
	}
	return nil
}

func NewHandler(appService service.IOrderApplicationService) OrderEventHandler {
	return &OrderEventHandlerImpl{orderAppService: appService}
}