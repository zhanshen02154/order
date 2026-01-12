package subscriber

import (
	"context"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/internal/domain/event/order"
	"github.com/zhanshen02154/order/internal/domain/event/product"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductEventHandler 商品事件处理器Handler
type ProductEventHandler interface {
	OnPaymentSuccessFailed(ctx context.Context, req *order.OnPaymentSuccess) error
	OnInventoryDeductSuccess(ctx context.Context, req *product.OnInventoryDeductSuccess) error
	RegisterSubscriber(srv server.Server)
}

// 商品事件处理器实现类
type productEventHandlerImpl struct {
	appSrv service.IOrderApplicationService
}

// NewProductEventHandler 创建商品事件处理器
func NewProductEventHandler(appSrv service.IOrderApplicationService) ProductEventHandler {
	return &productEventHandlerImpl{appSrv: appSrv}
}

// OnPaymentSuccessFailed 扣减库存失败后的处理
func (h *productEventHandlerImpl) OnPaymentSuccessFailed(ctx context.Context, req *order.OnPaymentSuccess) error {
	if req.OrderId == 0 {
		return status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}
	return h.appSrv.RevertPayStatus(ctx, req.OrderId)
}

// OnInventoryDeductSuccess 库存扣减成功
func (h *productEventHandlerImpl) OnInventoryDeductSuccess(ctx context.Context, req *product.OnInventoryDeductSuccess) error {
	if req.OrderId == 0 {
		return status.Error(codes.InvalidArgument, "orderId cannot be empty")
	}
	err := h.appSrv.ConfirmPayment(ctx, req.OrderId)
	return err
}

// OnInventoryDeductSuccessFailed 库存扣减失败
func (h *productEventHandlerImpl) OnInventoryDeductSuccessFailed(ctx context.Context, req *product.OnInventoryDeductSuccess) error {
	if req.OrderId == 0 {
		return status.Error(codes.InvalidArgument, "orderId cannot be empty")
	}
	err := h.appSrv.RevertPayStatus(ctx, req.OrderId)
	return err
}

// RegisterSubscriber 注册订阅者
func (h *productEventHandlerImpl) RegisterSubscriber(srv server.Server) {
	var err error
	queue := server.SubscriberQueue("order-consumer")
	err = micro.RegisterSubscriber("OnInventoryDeductSuccess", srv, h.OnInventoryDeductSuccess, queue)
	if err != nil {
		logger.Errorf("failed to register subscriber, error: %s", err.Error())
	}
	err = micro.RegisterSubscriber("OnPaymentSuccessDLQ", srv, h.OnPaymentSuccessFailed, queue)
	if err != nil {
		logger.Errorf("failed to register subscriber, error: %s", err.Error())
	}
	err = micro.RegisterSubscriber("OnInventoryDeductSuccessDLQ", srv, h.OnInventoryDeductSuccessFailed, queue)
	if err != nil {
		logger.Errorf("failed to register subscriber, error: %s", err.Error())
	}
}
