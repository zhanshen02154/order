package service

import (
	"git.imooc.com/zhanshen1614/order/internal/domain/model"
	"git.imooc.com/zhanshen1614/order/internal/domain/repository"
	order "git.imooc.com/zhanshen1614/order/proto/order"
	_ "time/tzdata"
)

type IOrderDataService interface {
	AddOrder(*model.Order) (int64, error)
	DeleteOrder(int64) error
	UpdateOrder(*model.Order) error
	FindOrderByID(int64) (*model.Order, error)
	GetOrderPagedList(page *order.OrderPageRequest) (*repository.Paginator[model.Order], error)
	UpdateShipStatus(int64, int32) error
	UpdatePayStatus(int64, int32) error
}

// 创建
func NewOrderDataService(orderRepository repository.IOrderRepository) IOrderDataService {
	return &OrderDataService{orderRepository}
}

type OrderDataService struct {
	OrderRepository repository.IOrderRepository
}

// 插入
func (u *OrderDataService) AddOrder(order *model.Order) (int64, error) {
	return u.OrderRepository.CreateOrder(order)
}

// 删除
func (u *OrderDataService) DeleteOrder(orderID int64) error {
	return u.OrderRepository.DeleteOrderByID(orderID)
}

// 更新
func (u *OrderDataService) UpdateOrder(order *model.Order) error {
	return u.OrderRepository.UpdateOrder(order)
}

// 根据id查找
func (u *OrderDataService) FindOrderByID(orderID int64) (*model.Order, error) {
	return u.OrderRepository.FindOrderByID(orderID)
}

// 查找
func (u *OrderDataService) GetOrderPagedList(page *order.OrderPageRequest) (*repository.Paginator[model.Order], error) {
	return u.OrderRepository.GetOrderPagedList(page)
}

func (u *OrderDataService) UpdateShipStatus(orderId int64, shipStatus int32) error {
	return u.OrderRepository.UpdateShipStatus(orderId, shipStatus)
}

func (u *OrderDataService) UpdatePayStatus(orderId int64, payStatus int32) error {
	return u.OrderRepository.UpdatePayStatus(orderId, payStatus)
}
