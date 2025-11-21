package handler

import (
	"context"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/pkg/swap"
	"github.com/zhanshen02154/order/proto/order"
	"go-micro.dev/v4/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
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
	err := o.OrderAppService.PayNotify(ctx, in)
	if err != nil {
		return errors.New(in.OutTradeNo, err.Error(), http.StatusPreconditionFailed)
	}
	resp.StatusCode = "0000"
	resp.Msg = "SUCCESS"
	return nil
}

func (o *OrderHandler) ConfirmPayment(ctx context.Context, in *order.PayNotifyRequest, out *order.ConfirmPaymentResponse) error {
	err := o.OrderAppService.ConfirmPayment(ctx, in)
	if err != nil {
		return errors.New(in.OutTradeNo, err.Error(), http.StatusPreconditionFailed)
	}
	out.Msg = "SUCCESS"
	out.StatusCode = "0000"
	out.Revert = false
	return nil
}

func (o *OrderHandler) ConfirmPaymentRevert(ctx context.Context, in *order.PayNotifyRequest, out *order.ConfirmPaymentResponse) error {
	err := o.OrderAppService.ConfirmPaymentRevert(ctx, in)
	if err != nil {
		return errors.New(in.OutTradeNo, err.Error(), http.StatusPreconditionFailed)
	}
	out.Msg = "SUCCESS"
	out.StatusCode = "0000"
	out.Revert = true
	return nil
}