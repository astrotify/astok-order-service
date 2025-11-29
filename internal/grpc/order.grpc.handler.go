package grpc

import (
	"context"
	"log"
	commonGrpc "order-service/go-proto/modules/common"
	orderGrpc "order-service/go-proto/modules/order"
	grpc "order-service/go-proto/services"
	"order-service/internal/database/db"
	"order-service/internal/service"
)

type OrderGrpcHandler struct {
	grpc.UnimplementedOrderGRPCServiceServer
	orderService *service.OrderService
}

func NewOrderGrpcHandler(orderService *service.OrderService) *OrderGrpcHandler {
	return &OrderGrpcHandler{
		orderService: orderService,
	}
}

func (h *OrderGrpcHandler) CreateOrder(ctx context.Context, req *orderGrpc.CreateOrderRequest) (*orderGrpc.CreateOrderResponse, error) {
	log.Printf("📥 Received CreateOrder request: UserID=%d, Products=%d", req.UserId, len(req.Products))

	products := make([]struct {
		ProductID int32
		Quantity  int32
		Price     float64
	}, len(req.Products))

	for i, p := range req.Products {
		products[i] = struct {
			ProductID int32
			Quantity  int32
			Price     float64
		}{
			ProductID: p.Id,
			Quantity:  p.Quantity,
			Price:     p.Price,
		}
	}

	order, orderProducts, err := h.orderService.CreateOrder(ctx, service.CreateOrderParams{
		UserID:      req.UserId,
		TotalAmount: req.TotalAmount,
		Products:    products,
	})

	if err != nil {
		log.Printf("❌ Failed to create order: %v", err)
		return &orderGrpc.CreateOrderResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &orderGrpc.CreateOrderResponse{
		Order: &orderGrpc.Order{
			Id:          order.ID,
			UserId:      order.UserID,
			TotalAmount: order.TotalAmount,
			Products:    h.mapProductsToProto(orderProducts),
		},
		Success: true,
		Message: "Order Created successfully",
	}, nil
}

func (h *OrderGrpcHandler) GetOrder(ctx context.Context, req *orderGrpc.GetOrderRequest) (*orderGrpc.GetOrderResponse, error) {
	log.Printf("📥 Received GetOrder request: ID=%d", req.Id)
	order, orderProducts, err := h.orderService.GetOrder(ctx, req.Id)

	if err != nil {
		return &orderGrpc.GetOrderResponse{
			Success: false,
			Message: "Order not found",
		}, nil
	}

	return &orderGrpc.GetOrderResponse{
		Success: true,
		Order: &orderGrpc.Order{
			Id:       order.ID,
			UserId:   order.UserID,
			Products: h.mapProductsToProto(orderProducts),
		},
		Message: "Order found",
	}, nil
}

func (h *OrderGrpcHandler) GetOrdersByUser(ctx context.Context, req *orderGrpc.GetOrdersByUserRequest) (*orderGrpc.GetOrdersByUserResponse, error) {
	log.Printf("📥 Received GetOrdersByUser request: UserID=%d", req.UserId)

	orders, total, totalPages, err := h.orderService.GetOrdersByUserId(ctx, req.UserId, req.Limit, req.Page)
	if err != nil {
		return &orderGrpc.GetOrdersByUserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	protoOrders := make([]*orderGrpc.Order, len(orders))
	for i, o := range orders {
		protoOrders[i] = &orderGrpc.Order{
			Id:          o.ID,
			UserId:      o.UserID,
			TotalAmount: o.TotalAmount,
		}
	}
	return &orderGrpc.GetOrdersByUserResponse{
		Orders: protoOrders,
		Pagination: &commonGrpc.Pagination{
			Total:      total,
			TotalPages: totalPages,
			Limit:      req.Limit,
			Page:       req.Page,
		},
		Success: true,
		Message: "Orders found",
	}, nil
}

func (h *OrderGrpcHandler) mapProductsToProto(products []db.OrderProduct) []*orderGrpc.OrderProduct {
	protoProducts := make([]*orderGrpc.OrderProduct, len(products))

	for i, p := range products {
		protoProducts[i] = &orderGrpc.OrderProduct{
			Price:    p.Price,
			Quantity: p.Quantity,
			Id:       p.ID,
		}
	}

	return protoProducts
}
