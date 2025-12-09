package handler

import (
	"context"
	"fmt"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/pkg/swap"
	"github.com/zhanshen02154/order/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderHandler struct {
	OrderAppService service.IOrderApplicationService
}

// 获取订单信息
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

// 支付回调
func (o *OrderHandler) PayNotify(ctx context.Context, in *order.PayNotifyRequest, resp *order.PayNotifyResponse) error {
	if in.OutTradeNo == "" {
		resp.StatusCode = "9999"
		resp.Msg = "OutTradeNo cannot be empty or null"
		return nil
	}
	err := o.OrderAppService.PayNotify(ctx, in)
	if err != nil {
		resp.StatusCode = "9999"
		resp.Msg = fmt.Sprintf("error to notify: %s", err.Error())
	}else {
		resp.StatusCode = "0000"
		resp.Msg = "SUCCESS"
	}
	return nil
}