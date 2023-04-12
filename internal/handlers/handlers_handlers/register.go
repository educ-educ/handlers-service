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

type registerHandlerOutDTO struct {
	HandlerID string `json:"handler_id"`
}

type handlerRegistrant interface {
	Register(specification handlers.Specification) (string, *http_tools.Error)
}

type RegisterHandler struct {
	logger   common.Logger
	service  handlerRegistrant
	validate *validator.Validate
}

func NewRegisterHandler(logger common.Logger, service handlerRegistrant, validate *validator.Validate) *RegisterHandler {
	return &RegisterHandler{
		logger:   logger,
		service:  service,
		validate: validate,
	}
}

func (handler *RegisterHandler) Handle(c *gin.Context) {
	handler.logger.Info("/handlers/register request received")

	var dto handlers.Specification
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

	handlerID, err := handler.service.Register(dto)
	if err != nil {
		_ = c.Error(err.AsGinError())
		return
	}

	c.JSON(http.StatusOK, registerHandlerOutDTO{HandlerID: handlerID})
}
