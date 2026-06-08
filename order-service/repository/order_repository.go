package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"order-service/model"
)

// OrderRepository defines the contract for order data access.
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	FindAll(ctx context.Context) ([]model.Order, error)
}

type orderRepository struct {
	collection *mongo.Collection
}

// NewOrderRepository creates a new OrderRepository backed by MongoDB.
func NewOrderRepository(db *mongo.Database) OrderRepository {
	return &orderRepository{
		collection: db.Collection("orders"),
	}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	order.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *orderRepository) FindAll(ctx context.Context) ([]model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find all orders: %w", err)
	}
	defer cursor.Close(ctx)

	orders := make([]model.Order, 0)
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("decode orders: %w", err)
	}

	return orders, nil
}
