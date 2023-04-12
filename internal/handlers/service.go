package handlers

import (
	"context"
	"github.com/educ-educ/handlers-service/internal/pkg/common"
	"github.com/educ-educ/handlers-service/internal/pkg/http_tools"
	"io"
	"net/http"
	"time"
)

type handlersRepo interface {
	GetUsedSockets() ([]string, *http_tools.Error)
	GetSpecification(handlerID string) (Specification, *http_tools.Error)
	AddHandlerInstance(socket string) (string, *http_tools.Error)
	AddMethods(handlerID string, methods []Method) *http_tools.Error
	RemoveHandler(handlerID string) *http_tools.Error
	UpdateHandler(handlerID string, specification Specification) *http_tools.Error
}

type handlersValidator interface {
	CheckHandler(specification Specification) *http_tools.Error
}

type Service struct {
	logger            common.Logger
	handlersRepo      handlersRepo
	handlersValidator handlersValidator
}

func NewService(logger common.Logger, handlersRepo handlersRepo, handlersValidator handlersValidator) *Service {
	return &Service{
		logger:            logger,
		handlersRepo:      handlersRepo,
		handlersValidator: handlersValidator,
	}
}

func (service *Service) GetSpecification(handlerID string) (Specification, *http_tools.Error) {
	return service.handlersRepo.GetSpecification(handlerID)
}

func (service *Service) Register(specification Specification) (string, *http_tools.Error) {
	isSocketUsed, httpErr := service.isSocketInUse(specification.Socket)
	if httpErr != nil {
		service.logger.Error(httpErr)
		return "", httpErr
	}
	if isSocketUsed {
		httpErr = &http_tools.Error{Type: http_tools.ValidationError, Info: "socket is already in use"}
		service.logger.Error(httpErr)
		return "", httpErr
	}

	if httpErr = service.handlersValidator.CheckHandler(specification); httpErr != nil {
		return "", httpErr
	}

	handlerID, httpErr := service.handlersRepo.AddHandlerInstance(specification.Socket)
	if httpErr != nil {
		return "", httpErr
	}

	if httpErr = service.handlersRepo.AddMethods(handlerID, specification.Methods); httpErr != nil {
		return "", httpErr
	}

	return handlerID, nil
}

func (service *Service) isSocketInUse(socket string) (bool, *http_tools.Error) {
	usedSockets, httpErr := service.handlersRepo.GetUsedSockets()
	if httpErr != nil {
		service.logger.Error(httpErr)
		return false, httpErr
	}
	return contains[string](usedSockets, socket), nil
}

func contains[V comparable](arr []V, val V) bool {
	for _, arrVal := range arr {
		if val == arrVal {
			return true
		}
	}
	return false
}

func (service *Service) Unregister(handlerID string) *http_tools.Error {
	return service.handlersRepo.RemoveHandler(handlerID)
}

func (service *Service) Update(handlerID string, specification Specification) *http_tools.Error {
	oldSpec, httpErr := service.handlersRepo.GetSpecification(handlerID)
	if httpErr != nil {
		service.logger.Error(httpErr)
		return httpErr
	}

	if oldSpec.Socket != specification.Socket {
		var isSocketUsed bool
		isSocketUsed, httpErr = service.isSocketInUse(specification.Socket)
		if httpErr != nil {
			service.logger.Error(httpErr)
			return httpErr
		}
		if isSocketUsed {
			httpErr = &http_tools.Error{Type: http_tools.ValidationError, Info: "socket is already in use"}
			service.logger.Error(httpErr)
			return httpErr
		}
	}

	if httpErr = service.handlersValidator.CheckHandler(specification); httpErr != nil {
		service.logger.Error(httpErr)
		return httpErr
	}

	if httpErr = service.handlersRepo.UpdateHandler(handlerID, specification); httpErr != nil {
		service.logger.Error(httpErr)
		return httpErr
	}

	return nil
}

func (service *Service) UseHandler(handlerID, path, method string, body io.Reader) (*http.Response, *http_tools.Error) {
	spec, httpErr := service.handlersRepo.GetSpecification(handlerID)
	if httpErr != nil {
		service.logger.Error(httpErr)
		return nil, httpErr
	}

	targetMethod := Method{MethodType: method, PathPart: path}
	if !contains[Method](spec.Methods, targetMethod) {
		httpErr = &http_tools.Error{Type: http_tools.NotFound, Info: "endpoint with given params is not found"}
		service.logger.Error(httpErr)
		return nil, httpErr
	}

	fullURL := spec.Socket + path

	reqCtx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	req, err := http.NewRequestWithContext(reqCtx, method, fullURL, body)
	if err != nil {
		httpErr = &http_tools.Error{Type: http_tools.NetworkError, Info: err.Error()}
		service.logger.Error(httpErr)
		return nil, httpErr
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		httpErr = &http_tools.Error{Type: http_tools.NetworkError, Info: err.Error()}
		service.logger.Error(httpErr)
		return nil, httpErr
	}

	return resp, nil
}
