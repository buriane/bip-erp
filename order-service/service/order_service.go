package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"order-service/model"
	"order-service/repository"
)

// Sentinel errors for the service layer — used by handler to map HTTP status codes.
var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidInput      = errors.New("invalid input")
)

// OrderService defines the contract for order business logic.
type OrderService interface {
	CreateOrder(ctx context.Context, req model.CreateOrderRequest) (*model.Order, error)
	GetAllOrders(ctx context.Context) ([]model.Order, error)
}

type orderService struct {
	repo              repository.OrderRepository
	productServiceURL string
	httpClient        *http.Client
}

// NewOrderService creates a new OrderService with the given repository and product service URL.
func NewOrderService(repo repository.OrderRepository, productServiceURL string) OrderService {
	return &orderService{
		repo:              repo,
		productServiceURL: productServiceURL,
		httpClient:        &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *orderService) CreateOrder(ctx context.Context, req model.CreateOrderRequest) (*model.Order, error) {
	if req.ProductID == "" {
		return nil, fmt.Errorf("%w: product id cannot be empty", ErrInvalidInput)
	}
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be greater than 0", ErrInvalidInput)
	}

	// Fetch product details from product-service (name, price snapshot)
	product, err := s.getProduct(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// Deduct stock via product-service internal endpoint
	if err := s.deductStock(ctx, req.ProductID, req.Quantity); err != nil {
		return nil, err
	}

	// Create order with snapshot data
	order := &model.Order{
		ProductID:   req.ProductID,
		ProductName: product.Name,
		Price:       product.Price,
		Quantity:    req.Quantity,
		TotalPrice:  product.Price * float64(req.Quantity),
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	orders, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all orders: %w", err)
	}

	return orders, nil
}

// productResponse is used to unmarshal the response from product-service GET /products/:id.
type productResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// deductStockPayload is the request body for POST /internal/deduct-stock.
type deductStockPayload struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// getProduct calls product-service to fetch product details.
// Maps HTTP status codes: 404 → ErrProductNotFound, 5xx → wrapped error.
func (s *orderService) getProduct(ctx context.Context, productID string) (*productResponse, error) {
	url := fmt.Sprintf("%s/products/%s", s.productServiceURL, productID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create get product request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call product service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrProductNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var product productResponse
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("decode product response: %w", err)
	}

	return &product, nil
}

// deductStock calls product-service internal endpoint to deduct stock atomically.
// Maps HTTP status codes: 404 → ErrProductNotFound, 400 → ErrInsufficientStock, 5xx → wrapped error.
func (s *orderService) deductStock(ctx context.Context, productID string, quantity int) error {
	url := fmt.Sprintf("%s/internal/deduct-stock", s.productServiceURL)

	payload := deductStockPayload{
		ProductID: productID,
		Quantity:  quantity,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal deduct stock request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create deduct stock request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call product service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrProductNotFound
	}
	if resp.StatusCode == http.StatusBadRequest {
		return ErrInsufficientStock
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	return nil
}
