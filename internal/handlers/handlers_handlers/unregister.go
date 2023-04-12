package handlers_handlers

import (
	"encoding/json"
	"github.com/educ-educ/handlers-service/internal/pkg/common"
	"github.com/educ-educ/handlers-service/internal/pkg/http_tools"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type unregisterHandlerDTO struct {
	HandlerID string `json:"handler_id" validate:"required"`
}

type handlerUnregistrant interface {
	Unregister(handlerID string) *http_tools.Error
}

type UnregisterHandler struct {
	logger   common.Logger
	service  handlerUnregistrant
	validate *validator.Validate
}

func NewUnregisterHandler(logger common.Logger, service handlerUnregistrant, validate *validator.Validate) *UnregisterHandler {
	return &UnregisterHandler{
		logger:   logger,
		service:  service,
		validate: validate,
	}
}

func (handler *UnregisterHandler) Handle(c *gin.Context) {
	handler.logger.Info("/handlers/unregister request received")

	var dto unregisterHandlerDTO
	if err := json.NewDecoder(c.Request.Body).Decode(&dto); err != nil {
		handler.logger.Error(err.Error())
		wrappedErr := http_tools.Error{Type: http_tools.ParseError, Info: err.Error()}
		_ = c.Error(wrappedErr.AsGinError())
		return
	}

	if err := handler.validate.Struct(dto); err != nil {
		handler.logger.Error(err.Error())
		for _, err := range err.(validator.ValidationErrors) {
			wrappedError := http_tools.Error{Type: http_tools.ValidationError, Info: err.Error()}
			_ = c.Error(wrappedError.AsGinError())
		}

		return
	}

	err := handler.service.Unregister(dto.HandlerID)
	if err != nil {
		_ = c.Error(err.AsGinError())
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}
