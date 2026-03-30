package subscriber

import (
	"context"
	"fmt"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/internal/domain/event/product"
	"github.com/zhanshen02154/order/internal/infrastructure/event"
	"go-micro.dev/v4/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductEventDlqHandler 商品事件处理器Handler
type ProductEventDlqHandler interface {
	OnInventoryDeductSuccessFailed(ctx context.Context, req *product.OnInventoryDeductSuccess) error
}

// 商品事件处理器死信主题实现类
type productEventDlqHandlerImpl struct {
	appSrv service.IOrderApplicationService
}

// OnInventoryDeductSuccessFailed 库存扣减失败
func (h *productEventDlqHandlerImpl) OnInventoryDeductSuccessFailed(ctx context.Context, req *product.OnInventoryDeductSuccess) error {
	if req.OrderId == 0 {
		return status.Error(codes.InvalidArgument, "orderId cannot be empty")
	}
	err := h.appSrv.RevertPayStatus(ctx, req.OrderId)
	return err
}

// AsEventHandlers 将 PaymentEventHandler 转换为 EventHandler 列表，用于注册到 EventDispatcher
func (h *productEventDlqHandlerImpl) AsEventHandlers() []event.EventHandler {
	return []event.EventHandler{
		event.NewGenericHandler("OnInventoryDeductSuccessDLQ", h.OnInventoryDeductSuccessFailed,
			func() *product.OnInventoryDeductSuccess {
				return &product.OnInventoryDeductSuccess{}
			},
		),
	}
}

// RegisterToDispatcher 注册到事件分发器
func (h *productEventDlqHandlerImpl) RegisterToDispatcher(dispatcher *event.EventDispatcher) error {
	handlers := h.AsEventHandlers()
	for _, handler := range handlers {
		if err := dispatcher.RegisterHandler(handler, "ProductEventDLQ", "order-consumer"); err != nil {
			return fmt.Errorf("failed to register handler %s: %w", handler.EventType(), err)
		} else {
			logger.Infof("registered product_dlq event handler: %s", handler.EventType())
		}
	}
	return nil
}

// NewProductDlqEventHandler 创建商品事件处理器
func NewProductDlqEventHandler(appSrv service.IOrderApplicationService) ProductEventDlqHandler {
	return &productEventDlqHandlerImpl{appSrv: appSrv}
}
