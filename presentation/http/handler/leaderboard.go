package handler

import (
	"github.com/hiamthach108/simplerank/internal/apperr"
	"github.com/hiamthach108/simplerank/internal/dto"
	"github.com/hiamthach108/simplerank/internal/service"
	"github.com/hiamthach108/simplerank/pkg/logger"
	"github.com/labstack/echo/v4"
)

type LeaderboardHandler struct {
	leaderboardSvc service.ILeaderboardSvc
	logger         logger.ILogger
}

func NewLeaderboardHandler(leaderboardSvc service.ILeaderboardSvc, logger logger.ILogger) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardSvc: leaderboardSvc,
		logger:         logger,
	}
}

func (h *LeaderboardHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/leaderboard/:id", h.GetLeaderboard)
	g.POST("/leaderboard/:id/score", h.SubmitScore)
}

func (h *LeaderboardHandler) GetLeaderboard(c echo.Context) error {
	leaderboardID := c.Param("id")
	scores, err := h.leaderboardSvc.GetTopEntries(c.Request().Context(), leaderboardID, 100)
	if err != nil {
		h.logger.Error("Failed to get leaderboard", "error", err)
		return HandleError(c, apperr.ErrLeaderboardNotFound, nil)
	}
	return HandleSuccess(c, scores)
}

func (h *LeaderboardHandler) SubmitScore(c echo.Context) error {
	leaderboardID := c.Param("id")
	var req dto.UpdateEntryScore
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err)
		return HandleError(c, apperr.ErrBadRequest, err)
	}
	if err := h.leaderboardSvc.UpdateEntryScore(c.Request().Context(), leaderboardID, req.EntryId, req.Score); err != nil {
		h.logger.Error("Failed to submit score", "error", err)
		return HandleError(c, apperr.ErrInternal, err)
	}
	return HandleSuccess(c, "Score submitted successfully")
}
