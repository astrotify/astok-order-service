package service

import (
	"context"
	"fmt"
	"log"
	"order-service/internal/database"
	"order-service/internal/database/db"
	"order-service/internal/database/kafka"
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
	TotalAmount int32
	Products    []struct {
		ProductID int32
		Quantity  int32
		Price     int32
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, params CreateOrderParams) (*db.Order, []db.OrderProduct, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	order, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
		UserID:      params.UserID,
		Status:      "PENDING",
		TotalAmount: params.TotalAmount,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order products
	products := make([]db.OrderProduct, len(params.Products))

	for _, p := range products {
		product, err := qtx.CreateOrderProduct(ctx, db.CreateOrderProductParams{
			ProductID: p.ID,
			OrderID:   p.OrderID,
			Quantity:  p.Quantity,
			Price:     p.Price,
		})

		if err != nil {
			return nil, nil, fmt.Errorf("failed to create order product: %w", err)
		}

		products = append(products, product)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
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
	} else {
		log.Printf("📨 Emitted order.created event to Kafka")
	}

	return &order, products, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderId int32) (*db.Order, []db.OrderProduct, error) {
	order, err := s.db.Queries.GetOrderByID(ctx, orderId)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get order: %w", err)
	}

	products, err := s.db.Queries.GetOrderProductsByOrderID(ctx, orderId)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get order products: %w", err)
	}

	return &order, products, nil
}

func (s *OrderService) GetOrdersByUserId(ctx context.Context, userId int32) ([]db.Order, error) {
	orders, err := s.db.Queries.GetOrdersByUserID(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderId int32, status string) (*db.Order, error) {
	order, err := s.db.Queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderId,
		Status: status,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}
	return &order, nil
}

func (s *OrderService) mapProductsToItems(products []db.OrderProduct) []map[string]interface{} {
	items := make([]map[string]interface{}, len(products))
	for i, p := range products {
		items[i] = map[string]interface{}{
			"productId": p.ProductID,
			"quantity":  p.Quantity,
		}
	}

	return items
}
