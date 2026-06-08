package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a pharmaceutical product in the system.
type Product struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Price     float64            `json:"price" bson:"price"`
	Stock     int                `json:"stock" bson:"stock"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateProductRequest is the payload for creating a new product.
type CreateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// UpdateProductRequest is the payload for updating a product.
// Pointer fields allow partial updates — only non-nil fields are applied.
type UpdateProductRequest struct {
	Name  *string  `json:"name"`
	Price *float64 `json:"price"`
	Stock *int     `json:"stock"`
}

// DeductStockRequest is the payload for the internal stock deduction endpoint.
type DeductStockRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}
