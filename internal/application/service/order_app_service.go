package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/service"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/proto/order"
)

type IOrderApplicationService interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	PayNotify(ctx context.Context, req *order.PayNotifyRequest) error
	ConfirmPayment(ctx context.Context, req *order.PayNotifyRequest) error
	ConfirmPaymentRevert(ctx context.Context, req *order.PayNotifyRequest) error
}

type OrderApplicationService struct {
	orderDataService service.IOrderDataService
	serviceContext   *infrastructure.ServiceContext
}

// 创建
func NewOrderApplicationService(
	serviceContext *infrastructure.ServiceContext,
) IOrderApplicationService {
	return &OrderApplicationService{
		orderDataService: service.NewOrderDataService(serviceContext.OrderRepository),
		serviceContext: serviceContext,
	}
}

// 根据ID获取订单信息
func (appService *OrderApplicationService) FindOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	return appService.orderDataService.FindOrderByID(ctx, id)
}

// 支付回调
func (appService *OrderApplicationService) PayNotify(ctx context.Context, req *order.PayNotifyRequest) error {
	lock, err := appService.serviceContext.LockManager.NewLock(ctx, fmt.Sprintf("orderpaynotify-%s", req.OutTradeNo), 60)
	defer lock.UnLock(ctx)
	if err != nil {
		return err
	}
	ok, err := lock.Lock(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("加锁失败")
	}

	orderInfo, err := appService.serviceContext.OrderRepository.FindPayOrderByCode(ctx, req.OutTradeNo)
	if err != nil {
		return err
	}
	if orderInfo.PayTime.Time.Unix() > 0 && orderInfo.PayStatus == 3 {
		return nil
	}
	addr := "127.0.0.1:9000"
	saga := appService.serviceContext.Dtm.BeginGrpcSaga(ctx).
		Add(addr + "/go.micro.service.Order/ConfirmPayment", addr + "/go.micro.service.Order/ConfirmPaymentRevert", req)
	saga.RetryLimit = 3
	saga.TimeoutToFail = 30
	err =  saga.Submit()
	return err
}

func (appService *OrderApplicationService) ConfirmPayment(ctx context.Context, req *order.PayNotifyRequest) error {
	return appService.serviceContext.TxManager.ExecuteWithBarrier(ctx, func(txCtx context.Context) error {
		return appService.orderDataService.ConfirmPayment(txCtx, req)
	})
}

func (appService *OrderApplicationService) ConfirmPaymentRevert(ctx context.Context, req *order.PayNotifyRequest) error {
	return appService.serviceContext.TxManager.ExecuteWithBarrier(ctx, func(txCtx context.Context) error {
		return appService.orderDataService.ConfirmPaymentRevert(txCtx, req)
	})
}

