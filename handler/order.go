package handler

import (
	"context"
	"git.imooc.com/zhanshen1614/common"
	"git.imooc.com/zhanshen1614/order/domain/model"
	"git.imooc.com/zhanshen1614/order/domain/service"
	order "git.imooc.com/zhanshen1614/order/proto/order"
)

type Order struct {
	OrderDataService service.IOrderDataService
}

func (o *Order) GetOrderById(ctx context.Context, request *order.OrderId, response *order.OrderInfo) error {
	orderInfo, err := o.OrderDataService.FindOrderByID(request.OrderId)
	if err != nil {
		return err
	}
	if err := common.SwapTo(orderInfo, response); err != nil {
		return err
	}
	return nil
}

func (o *Order) GetOrderPagedList(ctx context.Context, request *order.OrderPageRequest, response *order.OrderPagedList) error {
	pageList, err := o.OrderDataService.GetOrderPagedList(request)
	if err != nil {
		return err
	}
	response.Page = request.Page
	response.Pages = pageList.Pages
	response.Total = pageList.Total
	response.PageSize = pageList.PageSize
	for _, item := range pageList.Data {
		orderItem := &order.OrderInfo{}
		if err := common.SwapTo(item, orderItem); err != nil {
			return err
		}
		response.Data = append(response.Data, orderItem)
	}
	return nil
}

func (o *Order) CreateOrder(ctx context.Context, request *order.OrderInfo, response *order.OrderId) error {
	orderAdd := &model.Order{}
	if err := common.SwapTo(request, orderAdd); err != nil {
		return err
	}
	orderID, err := o.OrderDataService.AddOrder(orderAdd)
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
func (o *Order) DeleteOrderById(ctx context.Context, request *order.OrderId, response *order.Response) error {
	if err := o.OrderDataService.DeleteOrder(request.OrderId); err != nil {
		return err
	}
	response.Msg = "SUCCESS"
	return nil
}

func (o *Order) UpdateOrderPayStatus(ctx context.Context, request *order.PayStatus, response *order.Response) error {
	if err := o.OrderDataService.UpdatePayStatus(request.OrderId, request.PayStatus); err != nil {
		return err
	}
	response.Msg = "SUCCESS"
	return nil
}

//
// UpdateOrderShipStatus
//  @Description: 更新订单物流状态
//  @receiver o
//  @param ctx
//  @param request
//  @param response
//  @return error
//
func (o *Order) UpdateOrderShipStatus(ctx context.Context, request *order.ShipStatus, response *order.Response) error {
	if err := o.OrderDataService.UpdateShipStatus(request.OrderId, request.ShipStatus); err != nil {
		return err
	}
	response.Msg = "SUCCESS"
	return nil
}

func (o *Order) UpdateOrder(ctx context.Context, request *order.OrderInfo, response *order.Response) error {
	updateOrderInfo := model.Order{}
	if err := common.SwapTo(request, updateOrderInfo); err != nil {
		return err
	}
	if err := o.OrderDataService.UpdateOrder(&updateOrderInfo); err != nil {
		return err
	}
	response.Msg = "SUCCESS"
	return nil
}
