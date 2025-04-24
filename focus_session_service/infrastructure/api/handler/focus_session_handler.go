package handler

import (
	"focusspot/focussessionservice/application/dto"
	"focusspot/focussessionservice/application/usecases"
	"focusspot/focussessionservice/domain/entity"

	"github.com/gofiber/fiber/v2"
)

type FocusSessionHandler struct {
	sessionUseCase usecase.IFocusSessionUseCase
}

func NewFocusSessionHandler(sessionUseCase usecase.IFocusSessionUseCase) *FocusSessionHandler {
	return &FocusSessionHandler{
		sessionUseCase: sessionUseCase,
	}
}

func (h *FocusSessionHandler) CreateSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req dto.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	// TODO: Add proper validation

	session, err := h.sessionUseCase.CreateSession(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

func (h *FocusSessionHandler) GetSessionByID(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID := c.Params("id")

	session, err := h.sessionUseCase.GetSessionByID(c.Context(), sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(session)
}

func (h *FocusSessionHandler) GetUserSessions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	req := dto.GetSessionsRequest{
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
		Limit:     c.QueryInt("limit", 20),
		Offset:    c.QueryInt("offset", 0),
	}

	sessions, err := h.sessionUseCase.GetUserSessions(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(sessions)
}

func (h *FocusSessionHandler) GetActiveSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	session, err := h.sessionUseCase.GetActiveSession(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(session)
}

func (h *FocusSessionHandler) UpdateSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID := c.Params("id")

	var req dto.UpdateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	session, err := h.sessionUseCase.UpdateSession(c.Context(), sessionID, userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(session)
}

func (h *FocusSessionHandler) StartSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID := c.Params("id")

	session, err := h.sessionUseCase.StartSession(c.Context(), sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(session)
}

func (h *FocusSessionHandler) EndSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID := c.Params("id")

	var req dto.EndSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	session, err := h.sessionUseCase.EndSession(c.Context(), sessionID, userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(session)
}

func (h *FocusSessionHandler) CancelSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID := c.Params("id")

	session, err := h.sessionUseCase.CancelSession(c.Context(), sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(session)
}

func (h *FocusSessionHandler) DeleteSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID := c.Params("id")

	err := h.sessionUseCase.DeleteSession(c.Context(), sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Session deleted successfully",
	})
}

func (h *FocusSessionHandler) GetProductivityStats(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	req := dto.GetProductivityStatsRequest{
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}

	stats, err := h.sessionUseCase.GetProductivityStats(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}

func (h *FocusSessionHandler) GetProductivityTrends(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	req := dto.GetProductivityTrendsRequest{
		Period: entity.Period(entity.Weekly),
		Limit:  c.QueryInt("limit", 12),
	}

	trends, err := h.sessionUseCase.GetProductivityTrends(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(trends)
}
