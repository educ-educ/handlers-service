package handlers

import (
	"context"
	"github.com/educ-educ/handlers-service/internal/pkg/common"
	"github.com/educ-educ/handlers-service/internal/pkg/http_tools"
	"net/http"
	"time"
)

type Validator struct {
	logger common.Logger
}

func NewValidator(logger common.Logger) *Validator {
	return &Validator{logger: logger}
}

func (validator *Validator) CheckHandler(specification Specification) *http_tools.Error {
	reqCtx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	for _, method := range specification.Methods {
		req, err := http.NewRequestWithContext(reqCtx, method.MethodType, specification.Socket+method.PathPart, nil)
		if err != nil {
			httpErr := &http_tools.Error{Type: http_tools.NetworkError, Info: err.Error()}
			validator.logger.Error(httpErr)
			return httpErr
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			httpErr := &http_tools.Error{Type: http_tools.NetworkError, Info: err.Error()}
			validator.logger.Error(httpErr)
			return httpErr
		}
		if err = resp.Body.Close(); err != nil {
			httpErr := &http_tools.Error{Type: http_tools.NetworkError, Info: err.Error()}
			validator.logger.Error(httpErr)
		}

		if resp.StatusCode == http.StatusNotFound {
			httpErr := &http_tools.Error{Type: http_tools.ValidationError, Info: "given method is not found"}
			validator.logger.Error(httpErr)
			return httpErr
		}
	}

	return nil
}
