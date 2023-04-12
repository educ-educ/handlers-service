package handlers_handlers

import (
	"errors"
	"fmt"
	"github.com/educ-educ/handlers-service/internal/pkg/common"
	"github.com/educ-educ/handlers-service/internal/pkg/http_tools"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
)

type handlerProvider interface {
	UseHandler(handlerID, path, method string, body io.Reader) (*http.Response, *http_tools.Error)
}

type UseHandler struct {
	logger            common.Logger
	service           handlerProvider
	validate          *validator.Validate
	maxDataPerRequest int64
}

func NewUseHandler(logger common.Logger, service handlerProvider, validate *validator.Validate, maxDataPerRequest int64) *UseHandler {
	return &UseHandler{
		logger:            logger,
		service:           service,
		validate:          validate,
		maxDataPerRequest: maxDataPerRequest,
	}
}

func (handler *UseHandler) Handle(c *gin.Context) {
	handler.logger.Info("/handlers/use request received")

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, handler.maxDataPerRequest)

	if err := c.Request.ParseMultipartForm(handler.maxDataPerRequest); err != nil {
		handler.logger.Error(err)
		wrappedErr := http_tools.Error{Type: http_tools.ParseError, Info: err.Error()}
		_ = c.Error(wrappedErr.AsGinError())
		return
	}

	keys := []string{"handler_id", "path", "method"}
	mapValues := make(map[string]string, len(keys))
	for _, key := range keys {
		var httpErr *http_tools.Error
		if mapValues[key], httpErr = takeOneFromMultipart(c, key, "required"); httpErr != nil {
			return
		}
	}

	fileHeaders := c.Request.MultipartForm.File["body"]
	var body io.Reader
	if len(fileHeaders) > 0 {
		if len(fileHeaders) > 1 {
			err := errors.New("several inclusions of request body, one expected")
			handler.logger.Error(err)
			wrappedErr := http_tools.Error{Type: http_tools.FileHeaderOpenError, Info: err.Error()}
			_ = c.Error(wrappedErr.AsGinError())
			return
		}

		var err error
		body, err = fileHeaders[0].Open()
		if err != nil {
			handler.logger.Error(err)
			wrappedErr := http_tools.Error{Type: http_tools.FileHeaderOpenError, Info: err.Error()}
			_ = c.Error(wrappedErr.AsGinError())
			return
		}
	}

	response, httpErr := handler.service.UseHandler(mapValues[keys[0]], mapValues[keys[1]], mapValues[keys[2]], body)
	if httpErr != nil {
		_ = c.Error(httpErr.AsGinError())
		return
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			handler.logger.Error(err)
			wrappedErr := http_tools.Error{Type: http_tools.NetworkError, Info: err.Error()}
			_ = c.Error(wrappedErr.AsGinError())
		}
	}()

	c.Writer.WriteHeader(response.StatusCode)
	if _, err := io.Copy(c.Writer, response.Body); err != nil {
		handler.logger.Error(err)
	}
}

func takeOneFromMultipart(c *gin.Context, key, validationString string) (string, *http_tools.Error) {
	values := c.Request.MultipartForm.Value[key]
	if values == nil {
		wrappedErr := http_tools.Error{Type: http_tools.ParseError, Info: fmt.Sprint(key, " value must be provided")}
		_ = c.Error(wrappedErr.AsGinError())
		return "", &wrappedErr
	}

	if len(values) != 1 {
		wrappedErr := http_tools.Error{Type: http_tools.ParseError, Info: fmt.Sprint(key, " ")}
		_ = c.Error(wrappedErr.AsGinError())
		return "", &wrappedErr
	}

	if err := validator.New().Var(values[0], validationString); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			wrappedErr := http_tools.Error{Type: http_tools.ValidationError, Info: err.Error()}
			_ = c.Error(wrappedErr.AsGinError())
		}

		return "", &http_tools.Error{Type: http_tools.ValidationError, Info: err.Error()}
	}

	return values[0], nil
}
