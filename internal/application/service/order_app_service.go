package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/service"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/proto/order"
	"github.com/zhanshen02154/order/proto/product"
)

type IOrderApplicationService interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	PayNotify(ctx context.Context, req *order.PayNotifyRequest) error
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
	lock, err := appService.serviceContext.LockManager.NewLock(ctx, fmt.Sprintf("orderpaynotify:%s", req.OutTradeNo), 30)
	if err != nil {
		return err
	}
	ok, err := lock.Lock(ctx)
	defer lock.UnLock(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New(fmt.Sprintf("duplicate notify: %s", req.OutTradeNo))
	}

	orderInfo, err := appService.serviceContext.OrderRepository.FindPayOrderByCode(ctx, req.OutTradeNo)
	if err != nil {
		return err
	}
	if orderInfo.PayTime.Time.Unix() > 0 && orderInfo.PayStatus > 2 {
		return nil
	}
	err = appService.serviceContext.TxManager.Execute(ctx, func(txCtx context.Context) error {
		return appService.orderDataService.UpdateOrderPayStatus(txCtx, orderInfo, req)
	})
	if err != nil {
		return err
	}

	productReq := &product.OrderDetailReq{
		OrderId: orderInfo.Id,
		Products: []*product.ProductInvetoryItem{},
	}
	for _, item := range orderInfo.OrderDetail {
		productReq.Products = append(productReq.Products, &product.ProductInvetoryItem{
			ProductId:     item.ProductId,
			ProductNum:    item.ProductNum,
			ProductSizeId: item.ProductSizeId,
		})
	}
	productSvcAddr := appService.serviceContext.Conf.Consumer.Product.Addr
	saga := appService.serviceContext.Dtm.BeginGrpcSaga(ctx).
		Add(productSvcAddr + "/DeductInvetory",
		productSvcAddr + "/DeductInvetoryRevert", productReq)
	saga.TimeoutToFail = 30
	err =  saga.Submit()
	return err
}