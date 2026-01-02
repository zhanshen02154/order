package service

import (
	"context"
	"fmt"
	orderevent "github.com/zhanshen02154/order/internal/domain/event/order"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/service"
	"github.com/zhanshen02154/order/internal/infrastructure"
	"github.com/zhanshen02154/order/internal/infrastructure/event"
	"github.com/zhanshen02154/order/proto/order"
	"go-micro.dev/v4/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IOrderApplicationService interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	PayNotify(ctx context.Context, req *order.PayNotifyRequest) error
	RevertPayStatus(ctx context.Context, orderId int64) error
	ConfirmPayment(ctx context.Context, orderId int64) error
}

type OrderApplicationService struct {
	orderDataService service.IOrderDataService
	serviceContext   *infrastructure.ServiceContext
	eb               event.Listener
}

// 创建
func NewOrderApplicationService(
	serviceContext *infrastructure.ServiceContext,
	eb event.Listener,
) IOrderApplicationService {
	srv := &OrderApplicationService{
		orderDataService: service.NewOrderDataService(serviceContext.OrderRepository),
		serviceContext:   serviceContext,
		eb:               eb,
	}
	return srv
}

// 根据ID获取订单信息
func (appService *OrderApplicationService) FindOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	return appService.orderDataService.FindOrderByID(ctx, id)
}

// 支付回调
func (appService *OrderApplicationService) PayNotify(ctx context.Context, req *order.PayNotifyRequest) error {
	lock, err := appService.serviceContext.LockManager.NewLock(ctx, fmt.Sprintf("orderpaynotify-%s", req.OutTradeNo), 25)
	if err != nil {
		return err
	}
	err = lock.TryLock(ctx)
	defer func() {
		if err := lock.UnLock(ctx); err != nil {
			logger.Error("failed to unlock ", lock.GetKey(ctx), " error: ", err)
		}
	}()
	if err != nil {
		return err
	}

	orderInfo, err := appService.serviceContext.OrderRepository.FindPayOrderByCode(ctx, req.OutTradeNo)
	if err != nil {
		return err
	}
	if orderInfo.PayTime.Time.Unix() > 0 || orderInfo.PayStatus > 2 {
		return nil
	}

	return appService.serviceContext.TxManager.Execute(ctx, func(txCtx context.Context) error {
		// 标记为处理中
		err = appService.orderDataService.UpdateOrderPayStatus(txCtx, orderInfo.Id, 3)
		if err != nil {
			return err
		}

		// 发布事件
		onPaymentSuccessEvent := &orderevent.OnPaymentSuccess{
			OrderId:  orderInfo.Id,
			Products: make([]*orderevent.ProductInventoryItem, 0),
		}
		if orderInfo.OrderDetail != nil {
			for _, item := range orderInfo.OrderDetail {
				onPaymentSuccessEvent.Products = append(onPaymentSuccessEvent.Products, &orderevent.ProductInventoryItem{
					ProductId:     item.ProductId,
					ProductNum:    item.ProductNum,
					ProductSizeId: item.ProductSizeId,
				})
			}
			err = appService.eb.Publish(txCtx, "OnPaymentSuccess", onPaymentSuccessEvent, fmt.Sprintf("%d", orderInfo.Id))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// RevertPayStatus 恢复支付状态
func (appService *OrderApplicationService) RevertPayStatus(ctx context.Context, orderId int64) error {
	orderInfo, err := appService.orderDataService.FindByIdAndStatus(ctx, orderId, 3)
	if err != nil {
		return status.Errorf(codes.Internal, "order_id %d find error: ", orderId, err)
	}
	if orderInfo == nil {
		return status.Errorf(codes.Aborted, "order_id %d not found", orderId, err)
	}
	return appService.serviceContext.TxManager.Execute(ctx, func(txCtx context.Context) error {
		err := appService.orderDataService.FailedPayment(ctx, orderInfo)
		if err != nil {
			return status.Errorf(codes.Aborted, "failed to update status: %d", orderId)
		}
		return nil
	})
}

// ConfirmPayment 确认支付
func (appService *OrderApplicationService) ConfirmPayment(ctx context.Context, orderId int64) error {
	orderInfo, err := appService.orderDataService.FindByIdAndStatus(ctx, orderId, 3)
	if err != nil {
		return status.Errorf(codes.Internal, "order_id %d find error: ", orderId, err)
	}
	if orderInfo == nil {
		return status.Errorf(codes.Aborted, "order_id %d not found", orderId, err)
	}

	err = appService.serviceContext.TxManager.Execute(ctx, func(txCtx context.Context) error {
		err := appService.orderDataService.ConfirmPayment(txCtx, orderInfo)
		if err != nil {
			return status.Errorf(codes.Aborted, "failed to confirm payment on %d, error: %s", orderInfo.Id, err.Error())
		}
		return nil
	})
	return err
}
