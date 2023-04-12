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

type getSpecDTO struct {
	HandlerID string `json:"handler_id" validate:"required"`
}

type handlerSpecProvider interface {
	GetSpecification(handlerID string) (handlers.Specification, *http_tools.Error)
}

type GetSpecHandler struct {
	logger   common.Logger
	service  handlerSpecProvider
	validate *validator.Validate
}

func NewGetSpecHandler(logger common.Logger, service handlerSpecProvider, validate *validator.Validate) *GetSpecHandler {
	return &GetSpecHandler{
		logger:   logger,
		service:  service,
		validate: validate,
	}
}

func (handler *GetSpecHandler) Handle(c *gin.Context) {
	handler.logger.Info("/handlers/get-spec request received")

	var dto getSpecDTO
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

	spec, err := handler.service.GetSpecification(dto.HandlerID)
	if err != nil {
		_ = c.Error(err.AsGinError())
		return
	}

	c.JSON(http.StatusOK, spec)
}
