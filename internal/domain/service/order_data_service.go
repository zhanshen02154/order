package service

import (
	"context"
	"errors"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/repository"
)

type IOrderDataService interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	PayNotify(ctx context.Context, payOrderInfo *model.PayOrderInfo) error
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
func (u *OrderDataService) PayNotify(ctx context.Context, payOrderInfo *model.PayOrderInfo) error {
	orderModel := &model.Order{}
	if payOrderInfo.PayTime.Unix() > 0 {
		return errors.New("订单已支付")
	}
	err := u.orderRepository.UpdatePayOrder(ctx, orderModel)
	return err
}