package handler

import (
	"net/http"

	"github.com/hiamthach108/simplerank/internal/apperr"
	"github.com/labstack/echo/v4"
)

type BaseResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func HandleSuccess(c echo.Context, data any) error {
	resp := BaseResp{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	}
	return c.JSON(http.StatusOK, resp)
}

func HandleError(c echo.Context, code apperr.AppErrCode, err error) error {
	resp := BaseResp{
		Code:    int(code),
		Message: apperr.GetErrorMessage(int(code)),
	}
	if err != nil {
		resp.Message = err.Error()
	}
	if code >= 500 {
		return c.JSON(http.StatusInternalServerError, resp)
	}

	return c.JSON(int(code), resp)
}
