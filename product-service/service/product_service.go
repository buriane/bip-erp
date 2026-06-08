package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"product-service/model"
	"product-service/repository"
)

// Sentinel errors for the service layer — used by handler to map HTTP status codes.
var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidInput      = errors.New("invalid input")
)

// ProductService defines the contract for product business logic.
type ProductService interface {
	CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error)
	GetAllProducts(ctx context.Context) ([]model.Product, error)
	GetProductByID(ctx context.Context, id string) (*model.Product, error)
	UpdateProduct(ctx context.Context, id string, req model.UpdateProductRequest) (*model.Product, error)
	DeleteProduct(ctx context.Context, id string) error
	DeductStock(ctx context.Context, req model.DeductStockRequest) error
}

type productService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new ProductService with the given repository.
func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be greater than 0", ErrInvalidInput)
	}
	if req.Stock < 0 {
		return nil, fmt.Errorf("%w: stock cannot be negative", ErrInvalidInput)
	}

	now := time.Now()
	product := &model.Product{
		Name:      req.Name,
		Price:     req.Price,
		Stock:     req.Stock,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	return product, nil
}

func (s *productService) GetAllProducts(ctx context.Context) ([]model.Product, error) {
	products, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all products: %w", err)
	}

	return products, nil
}

func (s *productService) GetProductByID(ctx context.Context, id string) (*model.Product, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrProductNotFound
	}

	product, err := s.repo.FindByID(ctx, objID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by id: %w", err)
	}

	return product, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id string, req model.UpdateProductRequest) (*model.Product, error) {
	if req.Name != nil && *req.Name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
	}
	if req.Price != nil && *req.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be greater than 0", ErrInvalidInput)
	}
	if req.Stock != nil && *req.Stock < 0 {
		return nil, fmt.Errorf("%w: stock cannot be negative", ErrInvalidInput)
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrProductNotFound
	}

	product, err := s.repo.FindByID(ctx, objID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("update product: %w", err)
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	product.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, objID, product); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("update product: %w", err)
	}

	return product, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrProductNotFound
	}

	if err := s.repo.Delete(ctx, objID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProductNotFound
		}
		return fmt.Errorf("delete product: %w", err)
	}

	return nil
}

func (s *productService) DeductStock(ctx context.Context, req model.DeductStockRequest) error {
	if req.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than 0", ErrInvalidInput)
	}

	objID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		return ErrProductNotFound
	}

	if err := s.repo.DeductStock(ctx, objID, req.Quantity); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProductNotFound
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			return ErrInsufficientStock
		}
		return fmt.Errorf("deduct stock: %w", err)
	}

	return nil
}
