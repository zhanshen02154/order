package subscriber

import (
	"context"
	"fmt"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/internal/domain/event/order"
	"github.com/zhanshen02154/order/internal/infrastructure/event"
	"go-micro.dev/v4/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderDlqHandler interface {
	OnPaymentSuccessFailed(ctx context.Context, req *order.OnPaymentSuccess) error
}

// OrderDlqHandlerImpl 订单事件处理器死信主题实现类
type OrderDlqHandlerImpl struct {
	appSrv service.IOrderApplicationService
}

// OnPaymentSuccessFailed 扣减库存失败后的处理
func (h *OrderDlqHandlerImpl) OnPaymentSuccessFailed(ctx context.Context, req *order.OnPaymentSuccess) error {
	if req.OrderId == 0 {
		return status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}
	return h.appSrv.RevertPayStatus(ctx, req.OrderId)
}

// AsEventHandlers 将 OrderDlqHandlerImpl 转换为 EventHandler 列表，用于注册到 EventDispatcher
func (h *OrderDlqHandlerImpl) AsEventHandlers() []event.EventHandler {
	return []event.EventHandler{
		event.NewGenericHandler("OnPaymentSuccessDLQ", h.OnPaymentSuccessFailed,
			func() *order.OnPaymentSuccess {
				return &order.OnPaymentSuccess{}
			},
		),
	}
}

// RegisterToDispatcher 注册到事件分发器
func (h *OrderDlqHandlerImpl) RegisterToDispatcher(dispatcher *event.EventDispatcher) error {
	handlers := h.AsEventHandlers()
	for _, handler := range handlers {
		if err := dispatcher.RegisterHandler(handler, "OrderEventDLQ", "order-consumer"); err != nil {
			return fmt.Errorf("failed to register handler %s: %s", handler.EventType(), err.Error())
		} else {
			logger.Infof("registered order_dlq event handler: %s", handler.EventType())
		}
	}
	return nil
}

// NewOrderDlqEventHandler 创建订单死信主题事件处理器
func NewOrderDlqEventHandler(appSrv service.IOrderApplicationService) OrderDlqHandler {
	return &OrderDlqHandlerImpl{appSrv: appSrv}
}
