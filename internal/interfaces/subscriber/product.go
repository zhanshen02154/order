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

// ProductEventHandler 商品事件处理器Handler
type ProductEventHandler interface {
	OnInventoryDeductSuccess(ctx context.Context, req *product.OnInventoryDeductSuccess) error
}

// 商品事件处理器实现类
type productEventHandlerImpl struct {
	appSrv service.IOrderApplicationService
}

// NewProductEventHandler 创建商品事件处理器
func NewProductEventHandler(appSrv service.IOrderApplicationService) ProductEventHandler {
	return &productEventHandlerImpl{appSrv: appSrv}
}

// OnInventoryDeductSuccess 库存扣减成功
func (h *productEventHandlerImpl) OnInventoryDeductSuccess(ctx context.Context, req *product.OnInventoryDeductSuccess) error {
	if req.OrderId == 0 {
		return status.Error(codes.InvalidArgument, "orderId cannot be empty")
	}
	err := h.appSrv.ConfirmPayment(ctx, req.OrderId)
	return err
}

// AsEventHandlers 将 PaymentEventHandler 转换为 EventHandler 列表，用于注册到 EventDispatcher
func (h *productEventHandlerImpl) AsEventHandlers() []event.EventHandler {
	return []event.EventHandler{
		event.NewGenericHandler("OnInventoryDeductSuccess", h.OnInventoryDeductSuccess,
			func() *product.OnInventoryDeductSuccess {
				return &product.OnInventoryDeductSuccess{}
			},
		),
	}
}

// RegisterToDispatcher 注册到事件分发器
func (h *productEventHandlerImpl) RegisterToDispatcher(dispatcher *event.EventDispatcher) error {
	handlers := h.AsEventHandlers()
	for _, handler := range handlers {
		if err := dispatcher.RegisterHandler(handler, "ProductEvent", "order-consumer"); err != nil {
			return fmt.Errorf("failed to register handler %s: %w", handler.EventType(), err)
		}
		logger.Infof("registered product event handler: %s", handler.EventType())
	}
	return nil
}
