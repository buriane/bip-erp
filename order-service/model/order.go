package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order represents a customer order with product snapshot data.
type Order struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductID   string             `json:"productId" bson:"product_id"`
	ProductName string             `json:"productName" bson:"product_name"`
	Price       float64            `json:"price" bson:"price"`
	Quantity    int                `json:"quantity" bson:"quantity"`
	TotalPrice  float64            `json:"totalPrice" bson:"total_price"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
}

// CreateOrderRequest is the payload for creating a new order.
type CreateOrderRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}
