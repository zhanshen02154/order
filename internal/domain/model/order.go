package model

import (
	"database/sql"
	"time"
)

// Order 订单模型
type Order struct {
	Id          int64         `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	UserId      int64         `gorm:"type:bigint(20);not null;default:0;comment:'用户id'" json:"user_id"`
	OrderCode   string        `gorm:"type:varchar(100);unique_index;not null:default:''" json:"order_code"`
	TradeTime   time.Time     `gorm:"type:datetime;comment:'下单时间'" json:"trade_time"`
	PayStatus   int32         `gorm:"type:tinyint(4);not null;default:0;comment:'支付状态'" json:"pay_status"`
	ShipStatus  int32         `gorm:"type:tinyint(4);not null;default:0;comment:'运输状态'" json:"ship_status"`
	PayTime     sql.NullTime     `gorm:"type:datetime;comment:'支付时间'" json:"pay_time"`
	Price       float64       `gorm:"type:decimal(18,2);not null;default:0;comment:'支付金额'" json:"price"`
	OrderDetail []OrderDetail `gorm:"ForeignKey:OrderId" json:"order_detail"`
	CreatedAt   time.Time     `gorm:"type:datetime;comment:'创建时间'" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"type:datetime;comment:'更新时间'" json:"updated_at"`
}

type PayOrderInfo struct {
	Id int64 `json:"id"`
	PayStatus int32 `json:"pay_status"`
	PayTime   time.Time `json:"pay_time"`
	OrderDetailBasic []PayOrderDetailInfo `json:"order_detail_basic" gorm:"ForeignKey:OrderId"`
}
