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
	UpdateOrderPayStatus(ctx context.Context, orderId int64, status int32) error
	FindByIdAndStatus(ctx context.Context, orderId int64, status int32) (*model.Order, error)
	ConfirmPayment(ctx context.Context, orderInfo *model.Order) error
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

// 更新订单状态
func (u *OrderDataService) UpdateOrderPayStatus(ctx context.Context, orderId int64, status int32) error {
	return u.orderRepository.UpdatePayStatus(ctx, orderId, status)
}

// 根据ID和状态查找订单
func (u *OrderDataService) FindByIdAndStatus(ctx context.Context, orderId int64, status int32) (*model.Order, error) {
	return u.orderRepository.FindByIdAndStatus(ctx, orderId, status)
}

// 确认支付
func (u *OrderDataService) ConfirmPayment(ctx context.Context, orderInfo *model.Order) error {
	orderInfo.PayStatus = 4
	orderInfo.ShipStatus = 2
	orderInfo.PayTime = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	return u.orderRepository.ConfirmPayment(ctx, orderInfo)
}