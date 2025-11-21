package service

import (
	"context"
	"database/sql"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/repository"
	"github.com/zhanshen02154/order/proto/order"
	"time"
)

type IOrderDataService interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	PayNotify(ctx context.Context, payOrderInfo *model.Order, req *order.PayNotifyRequest) error
	ConfirmPayment(ctx context.Context, req *order.PayNotifyRequest) error
	ConfirmPaymentRevert(ctx context.Context, req *order.PayNotifyRequest) error
}

// 创建
func NewOrderDataService(orderRepository repository.IOrderRepository) IOrderDataService {
	return &OrderDataService{orderRepository: orderRepository}
}

type OrderDataService struct {
	orderRepository repository.IOrderRepository
}

// 根据id查找
func (u *OrderDataService) FindOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	return u.orderRepository.FindOrderByID(ctx, id)
}

// 订单支付回调
func (u *OrderDataService) PayNotify(ctx context.Context, payOrderInfo *model.Order, req *order.PayNotifyRequest) error {
	if req.StatusCode != "0000" {
		payOrderInfo.PayStatus = 4
	}
	err := u.orderRepository.UpdatePayOrder(ctx, payOrderInfo)
	return err
}

// 确认支付
func (u *OrderDataService) ConfirmPayment(ctx context.Context, req *order.PayNotifyRequest) error {
	orderInfo := &model.Order{
		OrderCode: req.OutTradeNo,
	}
	if req.StatusCode == "0000" {
		orderInfo.PayStatus = 3
		orderInfo.PayTime = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}
	err := u.orderRepository.ConfirmPaymentOrder(ctx, orderInfo)
	if err != nil {
		return err
	}
	return nil
}

// 确认支付的补偿操作
func (u *OrderDataService) ConfirmPaymentRevert(ctx context.Context, req *order.PayNotifyRequest) error {
	orderInfo := &model.Order{
		OrderCode: req.OutTradeNo,
		PayStatus: 2,
		PayTime: sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	}
	return u.orderRepository.ConfirmPaymentOrder(ctx, orderInfo)
}
