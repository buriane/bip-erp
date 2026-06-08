package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"product-service/model"
)

// Sentinel errors returned by the repository layer.
var (
	ErrNotFound          = errors.New("not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// ProductRepository defines the contract for product data access.
type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	FindAll(ctx context.Context) ([]model.Product, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Product, error)
	Update(ctx context.Context, id primitive.ObjectID, product *model.Product) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeductStock(ctx context.Context, id primitive.ObjectID, quantity int) error
}

type productRepository struct {
	collection *mongo.Collection
}

// NewProductRepository creates a new ProductRepository backed by MongoDB.
func NewProductRepository(db *mongo.Database) ProductRepository {
	return &productRepository{
		collection: db.Collection("products"),
	}
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}

	product.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *productRepository) FindAll(ctx context.Context) ([]model.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find all products: %w", err)
	}
	defer cursor.Close(ctx)

	products := make([]model.Product, 0)
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("decode products: %w", err)
	}

	return products, nil
}

func (r *productRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var product model.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find product by id: %w", err)
	}

	return &product, nil
}

func (r *productRepository) Update(ctx context.Context, id primitive.ObjectID, product *model.Product) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"name":       product.Name,
			"price":      product.Price,
			"stock":      product.Stock,
			"updated_at": product.UpdatedAt,
		},
	})
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *productRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

// DeductStock atomically decrements stock using FindOneAndUpdate.
// If no document matches (either product doesn't exist or stock is insufficient),
// a secondary FindOne determines the exact cause.
func (r *productRepository) DeductStock(ctx context.Context, id primitive.ObjectID, quantity int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":   id,
		"stock": bson.M{"$gte": quantity},
	}
	update := bson.M{
		"$inc": bson.M{"stock": -quantity},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result := r.collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			// Determine whether the product doesn't exist or stock is insufficient
			var existing model.Product
			err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&existing)
			if err != nil {
				return ErrNotFound
			}
			return ErrInsufficientStock
		}
		return fmt.Errorf("deduct stock: %w", result.Err())
	}

	return nil
}
