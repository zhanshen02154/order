package repository

import (
	"errors"
	"git.imooc.com/zhanshen1614/order/domain/model"
	order "git.imooc.com/zhanshen1614/order/proto/order"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

type IOrderRepository interface{
    InitTable() error
    FindOrderByID(int64) (*model.Order, error)
	CreateOrder(*model.Order) (int64, error)
	DeleteOrderByID(int64) error
	UpdateOrder(*model.Order) error
	GetOrderPagedList(*order.OrderPageRequest) (*Paginator[model.Order], error)
	UpdateShipStatus(int64, int32) error
	UpdatePayStatus(int64, int32) error
}

// NewOrderRepository 创建orderRepository
func NewOrderRepository(db *gorm.DB) IOrderRepository  {
	return &OrderRepository{mysqlDb:db}
}

type OrderRepository struct {
	mysqlDb *gorm.DB
}

// InitTable 初始化表
func (u *OrderRepository)InitTable() error  {
	if !u.mysqlDb.HasTable(&model.Order{}) && !u.mysqlDb.HasTable(&model.OrderDetail{}) {
		return u.mysqlDb.CreateTable(&model.Order{}, &model.OrderDetail{}).Error
	}
	return nil
}

// FindOrderByID 根据ID查找Order信息
func (u *OrderRepository)FindOrderByID(orderID int64) (order *model.Order,err error) {
	order = &model.Order{}
	return order, u.mysqlDb.Preload("OrderDetail").First(order,orderID).Error
}

// CreateOrder 创建Order信息
func (u *OrderRepository) CreateOrder(order *model.Order) (int64, error) {
	return order.Id, u.mysqlDb.Create(order).Error
}

// DeleteOrderByID 根据ID删除Order信息
func (u *OrderRepository) DeleteOrderByID(orderID int64) error {
	tx := u.mysqlDb.Begin()

	defer func() {
		if r := recover();r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Unscoped().Where("id = ?", orderID).Delete(&model.Order{}).Error;err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Where("order_id = ?", orderID).Delete(&model.OrderDetail{}).Error;err != nil {
		tx.Rollback()
		return err
	}
	return u.mysqlDb.Where("id = ?",orderID).Delete(&model.Order{}).Error
}

// UpdateOrder 更新Order信息
func (u *OrderRepository) UpdateOrder(order *model.Order) error {
	return u.mysqlDb.Model(order).Update(order).Error
}

// FindAll 获取结果集
func (u *OrderRepository) GetOrderPagedList(page *order.OrderPageRequest)(*Paginator[model.Order], error) {
	pageList := &Paginator[model.Order]{
		Page: page.Page,
		PageSize: page.PageSize,
	}
	var strBuilder strings.Builder
	strBuilder.WriteString("1=1")
	var queryArgs []interface{}
	if page.Conditions.OrderCode != "" {
		strBuilder.WriteString(" AND order_code = ?")
		queryArgs = append(queryArgs, page.Conditions.OrderCode)
	}
	if page.Conditions.PayStatus != 0 {
		strBuilder.WriteString(" AND pay_status = ?")
		queryArgs = append(queryArgs, page.Conditions.PayStatus)
	}
	if page.Conditions.ShipStatus != 0 {
		strBuilder.WriteString(" AND ship_status = ?")
		queryArgs = append(queryArgs, page.Conditions.ShipStatus)
	}
	if page.Conditions.OrderStartTime != nil {
		startTime := page.Conditions.OrderStartTime.AsTime()
		strBuilder.WriteString(" AND trade_time >= ?")
		queryArgs = append(queryArgs, startTime.Format("2006-01-02") + " 00:00:00")
	}
	if page.Conditions.OrderEndTime != nil {
		endTime := page.Conditions.OrderEndTime.AsTime()
		strBuilder.WriteString(" AND trade_time <= ?")
		queryArgs = append(queryArgs, endTime.Format("2006-01-02") + " 23:59:59")
	}

	query := u.mysqlDb.Model(&model.Order{}).Preload("OrderDetail").Where(strBuilder.String(), queryArgs...).Order("trade_time DESC")
	err := pageList.FindPagedList(query)
	return pageList, err
}

//
// UpdateShipStatus
//  @Description: 更新发货状态
//  @receiver u
//  @param orderID
//  @param shipStatus
//  @return error
//
func (u *OrderRepository) UpdateShipStatus(orderID int64, shipStatus int32) error {
	db := u.mysqlDb.Model(&model.Order{}).Where("id = ?", orderID).UpdateColumn("ship_status", shipStatus)
	if db.Error != nil {
		return db.Error
	}
	if db.RowsAffected == 0 {
		return errors.New("更新失败")
	}
	return nil
}

//
// UpdatePayStatus
//  @Description: 更新支付状态
//  @receiver u
//  @param orderID
//  @param payStatus
//  @return error
//
func (u *OrderRepository) UpdatePayStatus(orderID int64, payStatus int32) error {
	updateData := make(map[string]interface{})
	updateData["pay_status"] = payStatus
	if payStatus == 2 {
		updateData["pay_time"] = time.Now()
	}else if payStatus == 3 {
		updateData["pay_time"] = nil
	}
	db := u.mysqlDb.Model(&model.Order{}).Where("id = ?", orderID).Updates(updateData)
	if db.Error != nil {
		return db.Error
	}
	if db.RowsAffected == 0 {
		return errors.New("更新失败")
	}
	return nil
}
