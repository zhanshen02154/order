package model

import "time"

// OrderDetail 订单详情表
type OrderDetail struct {
	Id            int64     `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	ProductId     int64     `gorm:"type:bigint(20);not null;default:0;comment:'产品id'" json:"product_id"`
	ProductSizeId int64     `gorm:"type:int(11);not null;default:0;comment:'产品规格id'" json:"product_size_id"`
	ProductPrice  float64   `gorm:"type:decimal(18,2);not null;default:0;comment:'产品价格'" json:"product_price"`
	OrderId       int64     `gorm:"type:bigint(20);not null;default:0;comment'订单ID'" json:"order_id"`
	ProductNum    int64     `gorm:"type:int(11);not null;default:0;comment'产品数量'" json:"product_num"`
	CreatedAt     time.Time `gorm:"type:datetime;comment:'创建时间'" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:datetime;comment:'更新时间'" json:"updated_at"`
}

type PayOrderDetailInfo struct {
	ProductId     int64 `json:"product_id"`
	ProductNum    int64 `json:"product_num"`
	ProductSizeId int64 `json:"product_size_id"`
}
