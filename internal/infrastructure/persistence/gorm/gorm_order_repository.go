package gorm

import (
	"context"
	"errors"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.IOrderRepository {
	return &OrderRepository{db: db}
}

func (orderRepo *OrderRepository) FindOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	orderInfo := &model.Order{}
	return orderInfo, orderRepo.db.Preload("OrderDetail").First(orderInfo, id).Error
}

func (orderRepo *OrderRepository) FindPayOrderByCode(ctx context.Context, orderCode string) (*model.PayOrderInfo, error) {
	db := GetDBFromContext(ctx, orderRepo.db)
	payOrderInfo := &model.PayOrderInfo{}
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "order_code", "pay_status", "pay_time").Where("order_code = ?", orderCode).First(payOrderInfo).Error
	if err != nil {
		return nil, err
	}
	if payOrderInfo == nil {
		return nil, errors.New("订单不存在！")
	}
	err = db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", payOrderInfo.Id).Find(&payOrderInfo.OrderDetailBasic).Error
	if err != nil {
		return nil, err
	}
	return payOrderInfo, nil
}

// 更新订单状态
func (orderRepo *OrderRepository) UpdatePayOrder(ctx context.Context, orderInfo *model.Order) error {
	db := GetDBFromContext(ctx, orderRepo.db)
	return db.Model(model.Order{}).Where("id = ?", orderInfo.Id).Select("pay_status", "pay_time").Updates(model.Order{
		PayStatus: orderInfo.PayStatus,
		PayTime: orderInfo.PayTime,
	}).Error
}
