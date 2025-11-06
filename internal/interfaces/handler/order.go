package handler

import (
	"context"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/pkg/swap"
	"github.com/zhanshen02154/order/proto/order"
)

type OrderHandler struct {
	OrderAppService service.IOrderApplicationService
}

func (o *OrderHandler) GetOrderById(ctx context.Context, request *order.OrderId, response *order.OrderInfo) error {
	orderInfo, err := o.OrderAppService.FindOrderByID(ctx, request.OrderId)
	if err != nil {
		return err
	}
	if err = swap.ConvertTo(orderInfo, response); err != nil {
		return err
	}
	return nil
}

// 支付回调
func (o *OrderHandler) PayNotify(ctx context.Context, in *order.PayNotifyRequest, resp *order.PayNotifyResponse) error {
	err := o.OrderAppService.PayNotify(ctx, in)
	if err != nil {
		resp.StatusCode = "9999"
	}else {
		resp.StatusCode = "0000"
	}
	return nil
}
