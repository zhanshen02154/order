package service

import (
	"context"
	"errors"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/repository"
	"github.com/zhanshen02154/order/internal/domain/service"
	gorm2 "github.com/zhanshen02154/order/internal/infrastructure/persistence/gorm"
	"github.com/zhanshen02154/order/internal/infrastructure/persistence/transaction"
	"github.com/zhanshen02154/order/proto/order"
	"gorm.io/gorm"
)

type IOrderApplicationService interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	PayNotify(ctx context.Context, req *order.PayNotifyRequest) error
}

type OrderApplicationService struct {
	txManager        transaction.TransactionManager
	orderDataService service.IOrderDataService
	orderRepository  repository.IOrderRepository
}

// 创建
func NewOrderApplicationService(db *gorm.DB) IOrderApplicationService {
	orderRepo := gorm2.NewOrderRepository(db)
	return &OrderApplicationService{
		orderDataService: service.NewOrderDataService(orderRepo),
		orderRepository:  orderRepo,
		txManager: gorm2.NewGormTransactionManager(db),
	}
}

func (appService *OrderApplicationService) FindOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	return appService.orderDataService.FindOrderByID(ctx, id)
}

// 支付回调
func (appService *OrderApplicationService) PayNotify(ctx context.Context, req *order.PayNotifyRequest) error {
	if req.StatusCode != "0000" {
		return errors.New("支付失败")
	}
	return appService.txManager.ExecuteTransaction(ctx, func(ctx context.Context) error {
		orderInfo, err := appService.orderRepository.FindPayOrderByCode(ctx, req.OutTradeNo)
		if err != nil {
			return err
		}

		// 执行具体业务逻辑
		err = appService.orderDataService.PayNotify(ctx, orderInfo)
		if err != nil {
			return err
		}

		// 调用客户端


		return nil
	})
}