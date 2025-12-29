package gorm

import (
	"context"
	"errors"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/repository"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.IOrderRepository {
	return &OrderRepository{db: db}
}

func (orderRepo *OrderRepository) FindOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	db := GetDBFromContext(ctx, orderRepo.db)
	orderInfo := &model.Order{}
	db.Callback().Create()
	return orderInfo, db.Preload("OrderDetail").First(orderInfo, id).Error
}

func (orderRepo *OrderRepository) FindPayOrderByCode(ctx context.Context, orderCode string) (*model.Order, error) {
	payOrderInfo := &model.Order{}
	err := orderRepo.db.WithContext(ctx).Debug().Table("orders").Select("id", "order_code", "pay_status", "pay_time").Where("order_code = ?", orderCode).First(payOrderInfo).Error
	if err != nil {
		return nil, err
	}
	if payOrderInfo == nil {
		return nil, errors.New("订单不存在！")
	}
	err = orderRepo.db.WithContext(ctx).Debug().Table("order_details").
		Where("order_id = ?", payOrderInfo.Id).Select("product_id", "product_num", "product_size_id", "order_id").Find(&payOrderInfo.OrderDetail).Error
	if err != nil {
		return nil, err
	}
	return payOrderInfo, nil
}

// 更新订单状态
func (orderRepo *OrderRepository) UpdatePayOrder(ctx context.Context, orderInfo *model.Order) error {
	db, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		db = orderRepo.db.WithContext(ctx)
	}
	res := db.Debug().Model(orderInfo).Select("pay_time", "pay_status").Updates(orderInfo)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// 订单确认支付
func (orderRepo *OrderRepository) ConfirmPaymentOrder(ctx context.Context, orderInfo *model.Order) error {
	db, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		db = orderRepo.db.WithContext(ctx)
	}
	res := db.Debug().Model(orderInfo).Select("pay_status", "pay_time").Where("order_code = ?", orderInfo.OrderCode).Updates(model.Order{
		PayStatus: orderInfo.PayStatus,
		PayTime:   orderInfo.PayTime,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// 更新订单状态
func (orderRepo *OrderRepository) UpdatePayStatus(ctx context.Context, id int64, status int32) error {
	db, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		db = orderRepo.db.WithContext(ctx)
	}
	res := db.Model(model.Order{}).Where("id = ?", id).Select("pay_status").Update("pay_status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// 根据ID和状态查找订单内容
func (orderRepo *OrderRepository) FindByIdAndStatus(ctx context.Context, id int64, status int32) (*model.Order, error) {
	order := &model.Order{}
	err :=  orderRepo.db.WithContext(ctx).Model(order).Where("id = ? AND pay_status = ?", id, status).
		Select("id", "pay_status", "pay_time").First(order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}else {
			return nil, err
		}
	}
	return order, err
}

// 确认支付
func (orderRepo *OrderRepository) ConfirmPayment(ctx context.Context, orderInfo *model.Order) error {
	db := GetDBFromContext(ctx, orderRepo.db)
	res := db.Model(orderInfo).
		Select("pay_status", "ship_status", "pay_time").Updates(orderInfo)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}