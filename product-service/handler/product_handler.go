package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"product-service/model"
	"product-service/service"
)

// ProductHandler handles HTTP requests for product endpoints.
type ProductHandler struct {
	svc service.ProductService
}

// NewProductHandler creates a new ProductHandler with the given service.
func NewProductHandler(svc service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// RegisterRoutes registers all product routes on the Fiber app.
func (h *ProductHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/products", h.CreateProduct)
	app.Get("/products", h.GetAllProducts)
	app.Get("/products/:id", h.GetProductByID)
	app.Put("/products/:id", h.UpdateProduct)
	app.Delete("/products/:id", h.DeleteProduct)
	app.Post("/internal/deduct-stock", h.DeductStock)
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req model.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	product, err := h.svc.CreateProduct(c.Context(), req)
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

func (h *ProductHandler) GetAllProducts(c *fiber.Ctx) error {
	products, err := h.svc.GetAllProducts(c.Context())
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

func (h *ProductHandler) GetProductByID(c *fiber.Ctx) error {
	id := c.Params("id")

	product, err := h.svc.GetProductByID(c.Context(), id)
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(product)
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	product, err := h.svc.UpdateProduct(c.Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(product)
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.svc.DeleteProduct(c.Context(), id); err != nil {
		return mapError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "product deleted"})
}

func (h *ProductHandler) DeductStock(c *fiber.Ctx) error {
	var req model.DeductStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.DeductStock(c.Context(), req); err != nil {
		return mapError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "stock deducted"})
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
