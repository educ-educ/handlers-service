package main

import (
	"context"
	"fmt"
	"github.com/educ-educ/handlers-service/internal/handlers"
	"github.com/educ-educ/handlers-service/internal/handlers/handlers_handlers"
	"github.com/educ-educ/handlers-service/internal/pkg/http_tools"
	"github.com/educ-educ/handlers-service/internal/pkg/postgres"
	"github.com/educ-educ/handlers-service/internal/pkg/server"
	"github.com/go-playground/validator/v10"
	"log"
	"os"
	"path"

	"github.com/educ-educ/handlers-service/docs"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

const (
	Mb = 2 << 23
)

func main() {
	err := godotenv.Load("deploy_handlers_service/.env")
	if err != nil {
		fmt.Print(err)
		return
	}

	logConfig := zap.NewDevelopmentConfig()
	logConfig.DisableStacktrace = true
	baseLogger, err := logConfig.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func() {
		if err = baseLogger.Sync(); err != nil {
			log.Fatalf("can't flush log entities: %v", err)
		}
	}()

	logger := baseLogger.Sugar()

	dbContext, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()

	postgresDB, cancelDB, err := postgres.NewPostgresDb(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal("cannot open postgres connection")
	}
	defer cancelDB(postgresDB)

	validate := validator.New()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(http_tools.ErrorsMiddleware(logger, 5*Mb))

	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	storageFolder := "./files-storage"
	err = os.Mkdir(storageFolder, os.ModePerm) // must create before calling os.Open("./files-storage/*")
	if err != nil && !os.IsExist(err) {
		logger.Fatal(err)
	}

	sessionsFolder := path.Join(storageFolder, "sessions")
	err = os.Mkdir(sessionsFolder, os.ModePerm) // must create before calling os.Open("./sessions/*")
	if err != nil && !os.IsExist(err) {
		logger.Fatal(err)
	}

	handlersRouter := router.Group("/handlers")
	{
		handlersValidator := handlers.NewValidator(logger)
		handlersRepository := handlers.NewPostgresHandlersRepository(logger, dbContext, postgresDB)

		service := handlers.NewService(logger, handlersRepository, handlersValidator)

		getSpecHandler := handlers_handlers.NewGetSpecHandler(logger, service, validate)
		registerHandler := handlers_handlers.NewRegisterHandler(logger, service, validate)
		unregisterHandler := handlers_handlers.NewUnregisterHandler(logger, service, validate)
		updateHandler := handlers_handlers.NewUpdateHandler(logger, service, validate)
		useHandler := handlers_handlers.NewUseHandler(logger, service, validate, 5*Mb)

		handlersRouter.GET("/get-spec", getSpecHandler.Handle)
		handlersRouter.POST("/register", registerHandler.Handle)
		handlersRouter.DELETE("/unregister", unregisterHandler.Handle)
		handlersRouter.PUT("/update", updateHandler.Handle)
		handlersRouter.POST("/use", useHandler.Handle)
	}

	addr := ":" + os.Getenv("SERVICE_PORT")
	serv := server.NewServer(logger, router, addr)
	err = serv.Start()
	if err != nil {
		logger.Fatal(err)
	}
}
