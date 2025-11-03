package gorm

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/repository"
	order "github.com/zhanshen02154/order/proto/order"
	"strings"
	"time"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.IOrderRepository {
	return &OrderRepository{db: db}
}

func (orderRepo *OrderRepository) FindOrderByID(id int64) (*model.Order, error) {
	orderInfo := &model.Order{}
	return orderInfo, orderRepo.db.Preload("OrderDetail").First(orderInfo, id).Error
}

func (orderRepo *OrderRepository) CreateOrder(orderInfo *model.Order) (int64, error) {
	return orderInfo.Id, orderRepo.db.Create(orderInfo).Error
}

func (orderRepo *OrderRepository) DeleteOrderByID(id int64) error {
	tx := orderRepo.db.Begin()

	// 发生错误时
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Unscoped().Where("id = ?", id).Delete(&model.Order{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Unscoped().Where("order_id = ?", id).Delete(&model.OrderDetail{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return orderRepo.db.Where("id = ?", id).Delete(&model.Order{}).Error
}

func (orderRepo *OrderRepository) UpdateOrder(order *model.Order) error {
	return orderRepo.db.Model(order).Update(order).Error
}

func (orderRepo *OrderRepository) UpdateShipStatus(id int64, shipStatus int32) error {
	updateDb := orderRepo.db.Model(&model.Order{}).Where("id = ?", id).UpdateColumn("ship_status", shipStatus)
	if updateDb.Error != nil {
		return updateDb.Error
	}
	if updateDb.RowsAffected == 0 {
		return errors.New("更新失败")
	}
	return nil
}

func (orderRepo *OrderRepository) UpdatePayStatus(orderID int64, payStatus int32) error {
	updateData := make(map[string]interface{})
	updateData["pay_status"] = payStatus
	if payStatus == 2 {
		updateData["pay_time"] = time.Now()
	} else if payStatus == 3 {
		updateData["pay_time"] = nil
	}
	updateDb := orderRepo.db.Model(&model.Order{}).Where("id = ?", orderID).Updates(updateData)
	if updateDb.Error != nil {
		return updateDb.Error
	}
	if updateDb.RowsAffected == 0 {
		return errors.New("更新失败")
	}
	return nil
}

func (orderRepo *OrderRepository) GetOrderPagedList(findReq *order.OrderPageRequest) (*repository.Paginator[model.Order], error) {
	pageList := &repository.Paginator[model.Order]{
		Page:     findReq.Page,
		PageSize: findReq.PageSize,
	}
	var strBuilder strings.Builder
	strBuilder.WriteString("1=1")
	var queryArgs []interface{}
	if findReq.Conditions.OrderCode != "" {
		strBuilder.WriteString(" AND order_code = ?")
		queryArgs = append(queryArgs, findReq.Conditions.OrderCode)
	}
	if findReq.Conditions.PayStatus != 0 {
		strBuilder.WriteString(" AND pay_status = ?")
		queryArgs = append(queryArgs, findReq.Conditions.PayStatus)
	}
	if findReq.Conditions.ShipStatus != 0 {
		strBuilder.WriteString(" AND ship_status = ?")
		queryArgs = append(queryArgs, findReq.Conditions.ShipStatus)
	}
	if findReq.Conditions.OrderStartTime != nil {
		startTime := findReq.Conditions.OrderStartTime.AsTime()
		strBuilder.WriteString(" AND trade_time >= ?")
		queryArgs = append(queryArgs, startTime.Format("2006-01-02")+" 00:00:00")
	}
	if findReq.Conditions.OrderEndTime != nil {
		endTime := findReq.Conditions.OrderEndTime.AsTime()
		strBuilder.WriteString(" AND trade_time <= ?")
		queryArgs = append(queryArgs, endTime.Format("2006-01-02")+" 23:59:59")
	}

	query := orderRepo.db.Model(&model.Order{}).Preload("OrderDetail").Where(strBuilder.String(), queryArgs...).Order("trade_time DESC")
	err := pageList.FindPagedList(query)
	return pageList, err
}
