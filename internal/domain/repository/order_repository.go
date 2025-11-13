package repository

import (
	"context"
	"github.com/zhanshen02154/order/internal/domain/model"
)

type IOrderRepository interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	FindPayOrderByCode(ctx context.Context, orderCode string) (*model.Order, error)
	UpdatePayOrder(ctx context.Context, orderInfo *model.Order) error
}
