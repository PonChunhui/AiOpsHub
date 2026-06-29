package handler

import (
	"net/http"
	"strconv"

	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type BaseHandler struct{}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (h *BaseHandler) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Data: data,
	})
}

func (h *BaseHandler) Error(c *gin.Context, err error) {
	se := service.GetServiceError(err)

	statusCode := h.mapErrorCodeToHTTPStatus(se.Code)

	c.JSON(statusCode, Response{
		Code:    int(se.Code),
		Message: se.Message,
	})
}

func (h *BaseHandler) mapErrorCodeToHTTPStatus(code service.ErrorCode) int {
	switch code {
	case service.EntityNotFound, service.ToolNotFound, service.AgentNotFound:
		return http.StatusNotFound
	case service.InvalidParameter, service.MissingParameter:
		return http.StatusBadRequest
	case service.EntityAlreadyExists:
		return http.StatusConflict
	case service.InternalError, service.DatabaseError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (h *BaseHandler) GetPageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	return page, pageSize
}

func (h *BaseHandler) GetIDParam(c *gin.Context) string {
	return c.Param("id")
}

func (h *BaseHandler) BindJSON(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(req); err != nil {
		return service.NewServiceError(service.InvalidParameter, "请求参数错误", err)
	}
	return nil
}
