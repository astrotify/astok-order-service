package service

import (
	"context"
	"log"
	"math"
	"order-service/internal/database"
	"order-service/internal/database/db"
	"order-service/internal/errors"
	"order-service/internal/kafka"
	"time"
)

type OrderService struct {
	db       *database.DB
	producer *kafka.Producer
	topic    string
}

func NewOrderService(db *database.DB, producer *kafka.Producer, topic string) *OrderService {
	return &OrderService{
		db:       db,
		producer: producer,
		topic:    topic,
	}
}

type CreateOrderParams struct {
	UserID      int32
	TotalAmount float64
	Products    []struct {
		ProductID int32
		Quantity  int32
		Price     float64
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, params CreateOrderParams) (*db.Order, []db.OrderProduct, error) {
	// Validate input
	if params.UserID <= 0 {
		return nil, nil, errors.NewOrderError(errors.CodeInvalidInput, "user ID is required")
	}
	if len(params.Products) == 0 {
		return nil, nil, errors.NewOrderError(errors.CodeInvalidInput, "at least one product is required")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Printf("❌ Failed to begin transaction: %v", err)
		return nil, nil, errors.Wrap(errors.CodeDatabaseError, err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	order, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
		UserID:      params.UserID,
		Status:      "PENDING",
		TotalAmount: params.TotalAmount,
	})

	if err != nil {
		log.Printf("❌ Failed to create order: %v", err)
		return nil, nil, errors.NewOrderError(errors.CodeOrderCreateFailed, "failed to create order")
	}

	// Create order products
	products := make([]db.OrderProduct, len(params.Products))

	for i, p := range params.Products {
		product, err := qtx.CreateOrderProduct(ctx, db.CreateOrderProductParams{
			ProductID: p.ProductID,
			OrderID:   order.ID,
			Quantity:  p.Quantity,
			Price:     p.Price,
		})

		if err != nil {
			log.Printf("❌ Failed to create order product: %v", err)
			return nil, nil, errors.NewOrderError(errors.CodeOrderCreateFailed, "failed to create order product")
		}

		products[i] = product
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("❌ Failed to commit transaction: %v", err)
		return nil, nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	log.Printf("✅ Order created: ID=%d, UserID=%d", order.ID, order.UserID)

	event := map[string]interface{}{
		"orderId":     order.ID,
		"userId":      order.UserID,
		"status":      order.Status,
		"items":       s.mapProductsToItems(products),
		"totalAmount": order.TotalAmount,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if err = s.producer.Emit(s.topic, event); err != nil {
		log.Printf("⚠️ Failed to emit order.created event: %v", err)
		// Don't fail the order creation if Kafka fails
	} else {
		log.Printf("📨 Emitted order.created event to Kafka")
	}

	return &order, products, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderId int32) (*db.Order, []db.OrderProduct, error) {
	if orderId <= 0 {
		return nil, nil, errors.NewOrderError(errors.CodeInvalidInput, "order ID is required")
	}

	order, err := s.db.Queries.GetOrderByID(ctx, orderId)
	if err != nil {
		log.Printf("❌ Failed to get order: %v", err)
		return nil, nil, errors.ErrOrderNotFound
	}

	products, err := s.db.Queries.GetOrderProductsByOrderID(ctx, orderId)
	if err != nil {
		log.Printf("❌ Failed to get order products: %v", err)
		return nil, nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	return &order, products, nil
}

func (s *OrderService) GetOrdersByUserId(ctx context.Context, userId int32, limit int32, page int32) ([]db.Order, int32, int32, error) {
	if userId <= 0 {
		return nil, 0, 0, errors.NewOrderError(errors.CodeInvalidInput, "user ID is required")
	}
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if page <= 0 {
		page = 1 // Default page
	}

	orders, err := s.db.Queries.GetOrdersByUserID(ctx, db.GetOrdersByUserIDParams{
		UserID: userId,
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		log.Printf("❌ Failed to get orders: %v", err)
		return nil, 0, 0, errors.Wrap(errors.CodeDatabaseError, err)
	}

	total, err := s.db.Queries.GetOrdersByUserIDCount(ctx, userId)
	if err != nil {
		log.Printf("❌ Failed to get orders count: %v", err)
		return nil, 0, 0, errors.Wrap(errors.CodeDatabaseError, err)
	}

	totalPages := int32(math.Ceil(float64(total) / float64(limit)))

	return orders, int32(total), totalPages, nil
}

// Valid order statuses
var validStatuses = map[string]bool{
	"PENDING":    true,
	"CONFIRMED":  true,
	"PROCESSING": true,
	"SHIPPED":    true,
	"DELIVERED":  true,
	"CANCELLED":  true,
	"REFUNDED":   true,
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderId int32, status string) (*db.Order, error) {
	if orderId <= 0 {
		return nil, errors.NewOrderError(errors.CodeInvalidInput, "order ID is required")
	}
	if !validStatuses[status] {
		return nil, errors.NewOrderError(errors.CodeInvalidStatus, "invalid order status: "+status)
	}

	order, err := s.db.Queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderId,
		Status: status,
	})

	if err != nil {
		log.Printf("❌ Failed to update order status: %v", err)
		return nil, errors.ErrOrderUpdateFailed
	}

	log.Printf("✅ Order status updated: ID=%d, Status=%s", orderId, status)
	return &order, nil
}

func (s *OrderService) mapProductsToItems(products []db.OrderProduct) []map[string]interface{} {
	items := make([]map[string]interface{}, len(products))
	for i, p := range products {
		items[i] = map[string]interface{}{
			"productId": p.ProductID,
			"quantity":  p.Quantity,
			"price":     p.Price,
		}
	}

	return items
}
