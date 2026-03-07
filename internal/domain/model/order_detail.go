package model

import (
	"database/sql"
	"gorm.io/gorm"
	"time"
)

// OrderDetail 订单详情表模型
type OrderDetail struct {
	ID             int64          `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	OrderID        int64          `gorm:"type:bigint(20);not null;default:0;comment:'订单ID，关联order.id';index:idx_order_id" json:"order_id"`
	UserID         int64          `gorm:"type:bigint(20);not null;default:0;comment:'用户ID，冗余存储';index:idx_user_id" json:"user_id"`
	SkuID          int64          `gorm:"type:bigint(20);not null;default:0;comment:'SKU ID，关联商品微服务的product_sku.id';index:idx_sku_id" json:"sku_id"`
	ProductID      int64          `gorm:"type:bigint(20);not null;default:0;comment:'商品ID，冗余存储，关联商品微服务的product.id';index:idx_product_id" json:"product_id"`
	ProductNo      string         `gorm:"type:varchar(64);not null;default:'';comment:'商品编号，冗余存储'" json:"product_no"`
	SkuNo          string         `gorm:"type:varchar(64);not null;default:'';comment:'SKU编号，冗余存储';index:idx_sku_no" json:"sku_no"`
	ProductName    string         `gorm:"type:varchar(255);not null;default:'';comment:'商品名称，冗余存储（下单时的商品名称）'" json:"product_name"`
	SkuSpec        string         `gorm:"type:varchar(500);not null;default:'';comment:'SKU规格，冗余存储（下单时的规格）'" json:"sku_spec"`
	Quantity       uint32         `gorm:"type:int(11);not null;default:1;comment:'商品数量'" json:"quantity"`
	UnitPrice      float64        `gorm:"type:decimal(18,2);not null;default:0.00;comment:'商品单价（下单时的价格）'" json:"unit_price"`
	TotalPrice     float64        `gorm:"type:decimal(18,2);not null;default:0.00;comment:'商品总价 = quantity * unit_price'" json:"total_price"`
	DiscountAmount float64        `gorm:"type:decimal(18,2);not null;default:0.00;comment:'优惠金额'" json:"discount_amount"`
	ActualPayment  float64        `gorm:"type:decimal(18,2);not null;default:0.00;comment:'实付金额 = total_price - discount_amount'" json:"actual_payment"`
	ProductImage   sql.NullString `gorm:"type:varchar(500);comment:'商品图片，冗余存储（下单时的图片）'" json:"product_image"`
	RefundStatus   int8           `gorm:"type:tinyint(1);not null;default:0;comment:'退款状态：0-无退款 1-退款中 2-已退款';index:idx_refund_status" json:"refund_status"`
	RefundAmount   float64        `gorm:"type:decimal(18,2);not null;default:0.00;comment:'退款金额'" json:"refund_amount"`
	RefundReason   sql.NullString `gorm:"type:varchar(500);comment:'退款原因'" json:"refund_reason"`
	CommentStatus  int8           `gorm:"type:tinyint(1);not null;default:0;comment:'评价状态：0-未评价 1-已评价'" json:"comment_status"`
	CreatedAt      time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;comment:'创建时间';index:idx_created_at" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"type:datetime;not null;default:NULL;comment:'删除时间'" json:"deleted_at"`
}

// TableName 指定表名
func (*OrderDetail) TableName() string {
	return "order_details"
}
