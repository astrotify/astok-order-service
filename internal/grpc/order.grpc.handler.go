package grpc

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	commonGrpc "order-service/go-proto/modules/common"
	orderGrpc "order-service/go-proto/modules/order"
	grpc "order-service/go-proto/services"
	"order-service/internal/database/db"
	"order-service/internal/errors"
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
			ProductID: p.ProductId,
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
		orderErr := errors.GetError(err)
		log.Printf("❌ Failed to create order: %v", err)
		return &orderGrpc.CreateOrderResponse{
			Success: false,
			Message: orderErr.Message,
			Code:    orderErr.ErrorCode,
		}, nil
	}

	return &orderGrpc.CreateOrderResponse{
		Success: true,
		Message: "Order created successfully",
		Code:    "SUCCESS",
		Data:    orderToProto(order, orderProducts),
	}, nil
}

func (h *OrderGrpcHandler) GetOrder(ctx context.Context, req *orderGrpc.GetOrderRequest) (*orderGrpc.GetOrderResponse, error) {
	log.Printf("📥 Received GetOrder request: ID=%d", req.Id)
	order, orderProducts, err := h.orderService.GetOrder(ctx, req.Id)

	if err != nil {
		orderErr := errors.GetError(err)
		return &orderGrpc.GetOrderResponse{
			Success: false,
			Message: orderErr.Message,
			Code:    orderErr.ErrorCode,
		}, nil
	}

	return &orderGrpc.GetOrderResponse{
		Success: true,
		Message: "Order found",
		Code:    "SUCCESS",
		Data:    orderToProto(order, orderProducts),
	}, nil
}

func (h *OrderGrpcHandler) GetOrdersByUser(ctx context.Context, req *orderGrpc.GetOrdersByUserRequest) (*orderGrpc.GetOrdersByUserResponse, error) {
	log.Printf("📥 Received GetOrdersByUser request: UserID=%d", req.UserId)

	orders, total, totalPages, err := h.orderService.GetOrdersByUserId(ctx, req.UserId, req.Limit, req.Page)
	if err != nil {
		orderErr := errors.GetError(err)
		return &orderGrpc.GetOrdersByUserResponse{
			Success: false,
			Message: orderErr.Message,
			Code:    orderErr.ErrorCode,
		}, nil
	}

	protoOrders := make([]*orderGrpc.Order, len(orders))
	for i, o := range orders {
		protoOrders[i] = orderToProtoSimple(&o)
	}

	return &orderGrpc.GetOrdersByUserResponse{
		Success: true,
		Message: "Orders found",
		Code:    "SUCCESS",
		Data: &orderGrpc.GetOrdersByUserResponse_Data{
			Orders: protoOrders,
			Pagination: &commonGrpc.Pagination{
				Total:      total,
				TotalPages: totalPages,
				Limit:      req.Limit,
				Page:       req.Page,
			},
		},
	}, nil
}

func (h *OrderGrpcHandler) UpdateOrderStatus(ctx context.Context, req *orderGrpc.UpdateOrderStatusRequest) (*orderGrpc.UpdateOrderStatusResponse, error) {
	log.Printf("📥 Received UpdateOrderStatus request: ID=%d, Status=%s", req.Id, req.Status)

	order, err := h.orderService.UpdateOrderStatus(ctx, req.Id, req.Status)
	if err != nil {
		orderErr := errors.GetError(err)
		return &orderGrpc.UpdateOrderStatusResponse{
			Success: false,
			Message: orderErr.Message,
			Code:    orderErr.ErrorCode,
		}, nil
	}

	return &orderGrpc.UpdateOrderStatusResponse{
		Success: true,
		Message: "Order status updated successfully",
		Code:    "SUCCESS",
		Data:    orderToProtoSimple(order),
	}, nil
}

// Helper functions

func orderToProto(order *db.Order, products []db.OrderProduct) *orderGrpc.Order {
	return &orderGrpc.Order{
		Id:          order.ID,
		UserId:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		Products:    productsToProto(products),
		CreatedAt:   formatTimestamp(order.CreatedAt),
		UpdatedAt:   formatTimestamp(order.UpdatedAt),
	}
}

func orderToProtoSimple(order *db.Order) *orderGrpc.Order {
	return &orderGrpc.Order{
		Id:          order.ID,
		UserId:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		CreatedAt:   formatTimestamp(order.CreatedAt),
		UpdatedAt:   formatTimestamp(order.UpdatedAt),
	}
}

func productsToProto(products []db.OrderProduct) []*orderGrpc.OrderProduct {
	protoProducts := make([]*orderGrpc.OrderProduct, len(products))

	for i, p := range products {
		protoProducts[i] = &orderGrpc.OrderProduct{
			Id:        p.ID,
			ProductId: p.ProductID,
			Quantity:  p.Quantity,
			Price:     p.Price,
		}
	}

	return protoProducts
}

func formatTimestamp(t pgtype.Timestamp) string {
	if t.Valid {
		return t.Time.UTC().Format(time.RFC3339)
	}
	return ""
}
