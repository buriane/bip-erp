package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"order-service/model"
	"order-service/service"
)

// OrderHandler handles HTTP requests for order endpoints.
type OrderHandler struct {
	svc service.OrderService
}

// NewOrderHandler creates a new OrderHandler with the given service.
func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// RegisterRoutes registers all order routes on the Fiber app.
func (h *OrderHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/orders", h.CreateOrder)
	app.Get("/orders", h.GetAllOrders)
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req model.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	order, err := h.svc.CreateOrder(c.Context(), req)
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *OrderHandler) GetAllOrders(c *fiber.Ctx) error {
	orders, err := h.svc.GetAllOrders(c.Context())
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}

// mapError maps service-layer errors to appropriate HTTP responses.
func mapError(c *fiber.Ctx, err error) error {
	if errors.Is(err, service.ErrProductNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "product not found"})
	}
	if errors.Is(err, service.ErrInsufficientStock) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "insufficient stock"})
	}
	if errors.Is(err, service.ErrInvalidInput) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
