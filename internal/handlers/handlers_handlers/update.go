package handlers_handlers

import (
	"encoding/json"
	"github.com/educ-educ/handlers-service/internal/handlers"
	"github.com/educ-educ/handlers-service/internal/pkg/common"
	"github.com/educ-educ/handlers-service/internal/pkg/http_tools"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type updateDTO struct {
	HandlerID     string                 `json:"handler_id"`
	Specification handlers.Specification `json:"specification"`
}

type handlerUpdater interface {
	Update(handlerID string, specification handlers.Specification) *http_tools.Error
}

type UpdateHandler struct {
	logger   common.Logger
	service  handlerUpdater
	validate *validator.Validate
}

func NewUpdateHandler(logger common.Logger, service handlerUpdater, validate *validator.Validate) *UpdateHandler {
	return &UpdateHandler{
		logger:   logger,
		service:  service,
		validate: validate,
	}
}

func (handler *UpdateHandler) Handle(c *gin.Context) {
	handler.logger.Info("/handlers/update request received")

	var dto updateDTO
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

	err := handler.service.Update(dto.HandlerID, dto.Specification)
	if err != nil {
		_ = c.Error(err.AsGinError())
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}
