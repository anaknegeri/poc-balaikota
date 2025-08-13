package handler

import (
	"people-counting/internal/domain/entity"
	"people-counting/internal/domain/service"

	"github.com/gofiber/fiber/v2"
)

// AlertTypeHandler handles HTTP requests related to alert types
type AlertTypeHandler struct {
	alertTypeService service.AlertTypeService
}

// NewAlertTypeHandler creates a new alert type handler
func NewAlertTypeHandler(alertTypeService service.AlertTypeService) *AlertTypeHandler {
	return &AlertTypeHandler{
		alertTypeService: alertTypeService,
	}
}

// RegisterRoutes registers routes for this handler
func (h *AlertTypeHandler) RegisterRoutes(router fiber.Router) {
	alertTypes := router.Group("/alert-types")

	alertTypes.Get("/", h.ListAlertTypes)
	alertTypes.Post("/", h.CreateAlertType)
}

// ListAlertTypes handles getting all alert types
func (h *AlertTypeHandler) ListAlertTypes(c *fiber.Ctx) error {
	ctx := c.Context()

	alertTypes, err := h.alertTypeService.GetAllAlertTypes(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   "Error getting alert types: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"count": len(alertTypes),
		"data":  alertTypes,
	})
}

// CreateAlertType handles creating a new alert type
func (h *AlertTypeHandler) CreateAlertType(c *fiber.Ctx) error {
	ctx := c.Context()

	alertType := new(entity.AlertType)

	// Parse request body
	if err := c.BodyParser(alertType); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Invalid request body: " + err.Error(),
		})
	}

	// Create alert type
	_, err := h.alertTypeService.CreateAlertType(ctx, alertType)
	if err != nil {
		status := fiber.StatusInternalServerError

		// Check for validation errors
		if err.Error() == "name is required" ||
			err.Error() == "icon is required" ||
			err.Error() == "color is required" {
			status = fiber.StatusBadRequest
		}

		return c.Status(status).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"msg":   "Alert type created successfully",
		"data":  alertType,
	})
}
