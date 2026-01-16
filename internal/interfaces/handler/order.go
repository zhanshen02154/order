package handler

import (
	"context"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/pkg/swap"
	"github.com/zhanshen02154/order/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderHandler struct {
	OrderAppService service.IOrderApplicationService
}

// GetOrderById 获取订单信息
func (o *OrderHandler) GetOrderById(ctx context.Context, request *order.OrderId, response *order.OrderInfo) error {
	orderInfo, err := o.OrderAppService.FindOrderByID(ctx, request.OrderId)
	if err != nil {
		return status.Error(codes.DataLoss, err.Error())
	}
	if err = swap.ConvertTo(orderInfo, response); err != nil {
		return status.Error(codes.Aborted, err.Error())
	}
	return nil
}

// PayNotify 支付回调
func (o *OrderHandler) PayNotify(ctx context.Context, in *order.PayNotifyRequest, resp *order.PayNotifyResponse) error {
	if in.OutTradeNo == "" {
		return status.Error(codes.InvalidArgument, "outTradeNo cannot be empty")
	}
	err := o.OrderAppService.PayNotify(ctx, in)
	if err != nil {
		return status.Error(codes.Aborted, err.Error())
	}
	resp.StatusCode = "0000"
	resp.Msg = "SUCCESS"
	return nil
}
