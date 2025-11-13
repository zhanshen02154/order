package service

import (
	"context"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/repository"
	"github.com/zhanshen02154/order/proto/order"
	"time"
)

type IOrderDataService interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	PayNotify(ctx context.Context, payOrderInfo *model.Order, req *order.PayNotifyRequest) error
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
	if req.StatusCode == "0000" {
		payOrderInfo.PayStatus = 3
		payOrderInfo.PayTime = time.Now()
	}else {
		payOrderInfo.PayStatus = 4
	}
	err := u.orderRepository.UpdatePayOrder(ctx, payOrderInfo)
	return err
}