package handler

import (
	"github.com/hiamthach108/simplerank/internal/dto"
	"github.com/hiamthach108/simplerank/internal/errorx"
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
	g.GET("/:id", h.HandleGetLeaderboard)
	g.POST("/:id/score", h.HandleSubmitScore)
	g.GET("", h.HandleGetAllLeaderboards)
	g.POST("", h.HandleCreateLeaderboard)
	g.PUT("", h.HandleUpdateLeaderboard)
}

func (h *LeaderboardHandler) HandleGetLeaderboard(c echo.Context) error {
	reqCtx := c.Request().Context()

	leaderboardID := c.Param("id")
	leaderboard, err := h.leaderboardSvc.GetLeaderboardDetail(reqCtx, leaderboardID)
	if err != nil {
		h.logger.Error("Failed to get leaderboard", "error", err)
		return HandleError(c, errorx.ErrLeaderboardNotFound, nil)
	}

	return HandleSuccess(c, leaderboard)
}

func (h *LeaderboardHandler) HandleSubmitScore(c echo.Context) error {
	reqCtx := c.Request().Context()

	leaderboardID := c.Param("id")

	var req dto.UpdateEntryScore
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err)
		return HandleError(c, errorx.ErrBadRequest, err)
	}
	if err := h.leaderboardSvc.UpdateEntryScore(reqCtx, leaderboardID, req.EntryID, req.Score); err != nil {
		h.logger.Error("Failed to submit score", "error", err)
		return HandleError(c, errorx.ErrInternal, err)
	}

	return HandleSuccess(c, "Score submitted successfully")
}

func (h *LeaderboardHandler) HandleGetAllLeaderboards(c echo.Context) error {
	reqCtx := c.Request().Context()

	leaderboards, err := h.leaderboardSvc.GetListLeaderboards(reqCtx)
	if err != nil {
		h.logger.Error("Failed to get leaderboards", "error", err)
		return HandleError(c, errorx.ErrInternal, err)
	}

	return HandleSuccess(c, leaderboards)
}

func (h *LeaderboardHandler) HandleCreateLeaderboard(c echo.Context) error {
	reqCtx := c.Request().Context()

	var req dto.CreateLeaderboardReq
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err)
		return HandleError(c, errorx.ErrBadRequest, err)
	}
	leaderboard, err := h.leaderboardSvc.CreateLeaderboard(reqCtx, req)
	if err != nil {
		h.logger.Error("Failed to create leaderboard", "error", err)
		return HandleError(c, errorx.ErrInternal, err)
	}

	return HandleSuccess(c, leaderboard)
}

func (h *LeaderboardHandler) HandleUpdateLeaderboard(c echo.Context) error {
	reqCtx := c.Request().Context()

	var req dto.UpdateLeaderboardReq
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err)
		return HandleError(c, errorx.ErrBadRequest, err)
	}
	if err := h.leaderboardSvc.UpdateLeaderboard(reqCtx, req.ID, req); err != nil {
		h.logger.Error("Failed to update leaderboard", "error", err)
		return HandleError(c, errorx.ErrInternal, err)
	}

	return HandleSuccess(c, "Leaderboard updated successfully")
}
