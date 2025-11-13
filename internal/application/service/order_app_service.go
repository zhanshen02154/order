package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/internal/domain/repository"
	"github.com/zhanshen02154/order/internal/domain/service"
	"github.com/zhanshen02154/order/internal/infrastructure/persistence/transaction"
	"github.com/zhanshen02154/order/pkg/swap"
	"github.com/zhanshen02154/order/proto/order"
	"github.com/zhanshen02154/order/proto/product"
)

type IOrderApplicationService interface {
	FindOrderByID(ctx context.Context, id int64) (*model.Order, error)
	PayNotify(ctx context.Context, req *order.PayNotifyRequest) error
}

type OrderApplicationService struct {
	txManager        transaction.TransactionManager
	orderDataService service.IOrderDataService
	orderRepository  repository.IOrderRepository
	productServiceClient product.ProductService
}

// 创建
func NewOrderApplicationService(txManager transaction.TransactionManager, orderRepo repository.IOrderRepository, productSrv product.ProductService) IOrderApplicationService {
	return &OrderApplicationService{
		orderDataService: service.NewOrderDataService(orderRepo),
		orderRepository:  orderRepo,
		txManager: txManager,
		productServiceClient: productSrv,
	}
}

// 根据ID获取订单信息
func (appService *OrderApplicationService) FindOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	return appService.orderDataService.FindOrderByID(ctx, id)
}

// 支付回调
func (appService *OrderApplicationService) PayNotify(ctx context.Context, req *order.PayNotifyRequest) error {
	return appService.txManager.ExecuteTransaction(ctx, func(ctx context.Context) error {
		orderInfo, err := appService.orderRepository.FindPayOrderByCode(ctx, req.OutTradeNo)
		if err != nil {
			return err
		}
		if orderInfo.PayTime.Unix() > 0 && orderInfo.PayStatus == 3 {
			return nil
		}
		if len(orderInfo.OrderDetail) == 0 {
			return errors.New(fmt.Sprintf("pay notify error on order_id: %d: no details found", orderInfo.Id))
		}

		// 执行具体业务逻辑
		err = appService.orderDataService.PayNotify(ctx, orderInfo, req)
		if err != nil {
			return err
		}

		// 调用客户端
		productDetails := product.OrderDetailReq{
			OrderId: orderInfo.Id,
		}
		err = swap.ConvertTo(orderInfo.OrderDetail, &productDetails.Products)
		if err != nil {
			return err
		}
		resp, err := appService.productServiceClient.DeductInvetory(ctx, &productDetails)
		if err != nil {
			return err
		}
		if resp.StatusCode != "0000" {
			return errors.New("failed to deduct invetory")
		}

		return nil
	})
}