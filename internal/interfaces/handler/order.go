package handler

import (
	"context"
	"github.com/zhanshen02154/order/internal/application/service"
	"github.com/zhanshen02154/order/internal/domain/model"
	"github.com/zhanshen02154/order/pkg/swap"
	"github.com/zhanshen02154/order/proto/order"
)

type OrderHandler struct {
	OrderAppService service.IOrderApplicationService
}

func (o *OrderHandler) GetOrderById(ctx context.Context, request *order.OrderId, response *order.OrderInfo) error {
	orderInfo, err := o.OrderAppService.FindOrderByID(request.OrderId)
	if err != nil {
		return err
	}
	if err := swap.ConvertTo(orderInfo, response); err != nil {
		return err
	}
	return nil
}

func (o *OrderHandler) GetOrderPagedList(ctx context.Context, request *order.OrderPageRequest, response *order.OrderPagedList) error {
	pageList, err := o.OrderAppService.GetOrderPagedList(request)
	if err != nil {
		return err
	}
	response.Page = request.Page
	response.Pages = pageList.Pages
	response.Total = pageList.Total
	response.PageSize = pageList.PageSize
	for _, item := range pageList.Data {
		orderItem := &order.OrderInfo{}
		if err := swap.ConvertTo(item, orderItem); err != nil {
			return err
		}
		response.Data = append(response.Data, orderItem)
	}
	return nil
}

func (o *OrderHandler) CreateOrder(ctx context.Context, request *order.OrderInfo, response *order.OrderId) error {
	orderAdd := &model.Order{}
	if err := swap.ConvertTo(request, orderAdd); err != nil {
		return err
	}
	orderID, err := o.OrderAppService.AddOrder(orderAdd)
	if err != nil {
		return err
	}
	response.OrderId = orderID
	return nil
}

// DeleteOrderById
//
//	@Description: 根据orderid删除订单
//	@receiver o
//	@param ctx
//	@param request
//	@param response
//	@return error
func (o *OrderHandler) DeleteOrderById(ctx context.Context, request *order.OrderId, response *order.Response) error {
	if err := o.OrderAppService.DeleteOrder(request.OrderId); err != nil {
		return err
	}
	response.Msg = "SUCCESS"
	return nil
}

func (o *OrderHandler) UpdateOrderPayStatus(ctx context.Context, request *order.PayStatus, response *order.Response) error {
	if err := o.OrderAppService.UpdatePayStatus(request.OrderId, request.PayStatus); err != nil {
		return err
	}
	response.Msg = "SUCCESS"
	return nil
}

// UpdateOrderShipStatus
//
//	@Description: 更新订单物流状态
//	@receiver o
//	@param ctx
//	@param request
//	@param response
//	@return error
func (o *OrderHandler) UpdateOrderShipStatus(ctx context.Context, request *order.ShipStatus, response *order.Response) error {
	if err := o.OrderAppService.UpdateShipStatus(request.OrderId, request.ShipStatus); err != nil {
		return err
	}
	response.Msg = "SUCCESS"
	return nil
}

func (o *OrderHandler) UpdateOrder(ctx context.Context, request *order.OrderInfo, response *order.Response) error {
	updateOrderInfo := model.Order{}
	if err := swap.ConvertTo(request, updateOrderInfo); err != nil {
		return err
	}
	if err := o.OrderAppService.UpdateOrder(&updateOrderInfo); err != nil {
		return err
	}
	response.Msg = "SUCCESS"
	return nil
}
