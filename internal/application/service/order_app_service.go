package service

import (
	"order/internal/domain/model"
	"order/internal/domain/repository"
	"order/internal/domain/service"
	order "order/proto/order"
)

type IOrderApplicationService interface {
	AddOrder(orderInfo *model.Order) (int64, error)
	DeleteOrder(id int64) error
	UpdateOrder(orderInfo *model.Order) error
	FindOrderByID(id int64) (*model.Order, error)
	GetOrderPagedList(page *order.OrderPageRequest) (*repository.Paginator[model.Order], error)
	UpdateShipStatus(id int64, status int32) error
	UpdatePayStatus(id int64, status int32) error
}

type OrderApplicationService struct {
	orderDataService service.IOrderDataService
	orderRepository  repository.IOrderRepository
}

// 创建
func NewOrderApplicationService(orderRepo repository.IOrderRepository) IOrderApplicationService {
	orderDataService := service.NewOrderDataService(orderRepo)
	return &OrderApplicationService{
		orderDataService: orderDataService,
		orderRepository:  orderRepo,
	}
}

func (appService *OrderApplicationService) AddOrder(orderInfo *model.Order) (int64, error) {
	return appService.orderDataService.Add(orderInfo)
}

func (appService *OrderApplicationService) DeleteOrder(id int64) error {
	return appService.orderDataService.Delete(id)
}

func (appService *OrderApplicationService) UpdateOrder(orderInfo *model.Order) error {
	return appService.orderDataService.Update(orderInfo)
}

func (appService *OrderApplicationService) FindOrderByID(id int64) (*model.Order, error) {
	return appService.orderDataService.FindOrderByID(id)
}

func (appService *OrderApplicationService) GetOrderPagedList(page *order.OrderPageRequest) (*repository.Paginator[model.Order], error) {
	return appService.orderDataService.GetOrderPagedList(page)
}

func (appService *OrderApplicationService) UpdateShipStatus(id int64, status int32) error {
	return appService.orderDataService.UpdateShipStatus(id, status)
}

func (appService *OrderApplicationService) UpdatePayStatus(id int64, status int32) error {
	return appService.orderDataService.UpdatePayStatus(id, status)
}
