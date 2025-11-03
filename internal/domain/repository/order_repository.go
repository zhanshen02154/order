package repository

import (
	"order/internal/domain/model"
	order "order/proto/order"
)

type IOrderRepository interface {
	FindOrderByID(int64) (*model.Order, error)
	CreateOrder(orderInfo *model.Order) (int64, error)
	DeleteOrderByID(int64) error
	UpdateOrder(*model.Order) error
	GetOrderPagedList(findReq *order.OrderPageRequest) (*Paginator[model.Order], error)
	UpdateShipStatus(int64, int32) error
	UpdatePayStatus(int64, int32) error
}
