package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	apiv1 "git.iu7.bmstu.ru/vai20u117/testing/src/internal/api/v1/controller"
	apiv2 "git.iu7.bmstu.ru/vai20u117/testing/src/internal/api/v2/controller"
	dbpostgres "git.iu7.bmstu.ru/vai20u117/testing/src/internal/db/postgres"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"
	_ "git.iu7.bmstu.ru/vai20u117/testing/src/swagger"
	muxhandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	requestsTimeout = 15 * time.Second

	adminSecretEnv = "ADMIN_SECRET"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetLogLoggerLevel(slog.LevelInfo)

	mustLoadConfigs()

	choice := mustGetChoice()

	database := mustLoadDB(ctx, choice)
	defer database.GetPool(ctx).Close()

	controllerV1 := createControllerV1(database)
	controllerV2 := createControllerV2(database)

	serverPort := ":" + getAppPort(choice)
	router := mux.NewRouter()
	// controller.CreateRouter(router)
	controllerV1.CreateRouter(router.PathPrefix("/api/v1").Subrouter())
	controllerV2.CreateRouter(router.PathPrefix("/api/v2").Subrouter())

	server := &http.Server{
		Addr:         serverPort,
		ReadTimeout:  requestsTimeout,
		WriteTimeout: requestsTimeout,
		Handler:      muxhandlers.CompressHandler(router),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("Failed to start server: ", err)
		}
	}()

	slog.Info("Server started", "address", "http://localhost"+serverPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("Server shutting down")
}

func createControllerV1(database *dbpostgres.Database) *apiv1.Controller {
	return apiv1.NewController(
		initPosterHandlerV1(database),
		initListHandlerV1(database),
		initListPosterHandlerV1(database),
		initPosterRecordHandlerV1(database),
		initAuthHandlerV1(database, os.Getenv(adminSecretEnv)),
	)
}

func createControllerV2(database *dbpostgres.Database) *apiv2.Controller {
	return apiv2.NewController(
		initPosterHandlerV2(database),
		initListHandlerV2(database),
		initListPosterHandlerV2(database),
		initPosterRecordHandlerV2(database),
		initAuthHandlerV2(database, os.Getenv(adminSecretEnv)),
	)
}

func initPosterHandlerV1(database *dbpostgres.Database) *apiv1.PosterHandler {
	posterRepository := repository.NewPosterRepository(database)
	posterService := service.NewPosterService(posterRepository)
	return apiv1.NewPosterHandler(posterService)
}

func initListHandlerV1(database *dbpostgres.Database) *apiv1.ListHandler {
	listRepository := repository.NewListRepository(database)
	listService := service.NewListService(listRepository)
	return apiv1.NewListHandler(listService)
}

func initListPosterHandlerV1(database *dbpostgres.Database) *apiv1.ListPosterHandler {
	listPosterRepository := repository.NewListPosterRepository(database)
	listPosterService := service.NewListPosterService(listPosterRepository)
	return apiv1.NewListPosterHandler(listPosterService)
}

func initPosterRecordHandlerV1(database *dbpostgres.Database) *apiv1.PosterRecordHandler {
	historyRepository := repository.NewPosterRecordRepository(database)
	historyService := service.NewPosterRecordService(historyRepository)
	return apiv1.NewPosterRecordHandler(historyService)
}

func initAuthHandlerV1(database *dbpostgres.Database, adminToken string) *apiv1.AuthHandler {
	userRepository := repository.NewUserRepository(database)
	authService := service.NewAuthService(userRepository, adminToken)
	return apiv1.NewAuthHandler(authService)
}

func initPosterHandlerV2(database *dbpostgres.Database) *apiv2.PosterHandler {
	posterRepository := repository.NewPosterRepository(database)
	posterService := service.NewPosterService(posterRepository)
	return apiv2.NewPosterHandler(posterService)
}

func initListHandlerV2(database *dbpostgres.Database) *apiv2.ListHandler {
	listRepository := repository.NewListRepository(database)
	listService := service.NewListService(listRepository)
	return apiv2.NewListHandler(listService)
}

func initListPosterHandlerV2(database *dbpostgres.Database) *apiv2.ListPosterHandler {
	listPosterRepository := repository.NewListPosterRepository(database)
	listPosterService := service.NewListPosterService(listPosterRepository)
	return apiv2.NewListPosterHandler(listPosterService)
}

func initPosterRecordHandlerV2(database *dbpostgres.Database) *apiv2.PosterRecordHandler {
	historyRepository := repository.NewPosterRecordRepository(database)
	historyService := service.NewPosterRecordService(historyRepository)
	return apiv2.NewPosterRecordHandler(historyService)
}

func initAuthHandlerV2(database *dbpostgres.Database, adminToken string) *apiv2.AuthHandler {
	userRepository := repository.NewUserRepository(database)
	authService := service.NewAuthService(userRepository, adminToken)
	return apiv2.NewAuthHandler(authService)
}

func mustLoadConfigs() {
	if err := initConfig(); err != nil {
		log.Fatal("Failed to init configs: ", err)
	}
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load env variables: ", err)
	}
}

func mustLoadDB(ctx context.Context, choice int) *dbpostgres.Database {
	database, err := dbpostgres.NewDB(ctx, &dbpostgres.DBConfig{
		Host: viper.GetString("db.host"),
		// Port:     viper.GetString("db.port"),
		Port:     getDBPort(choice),
		Username: getDBUsername(choice),
		Password: getDBPassword(choice),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		log.Fatal("Failed to create db: ", err)
	}

	return database
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func mustGetChoice() int {
	choice, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("failed to get choid, err: ", err)
	}

	return choice
}

func getDBUsername(choice int) string {
	switch choice {
	case 1:
		return os.Getenv("DB_USERNAME_RW")
	default:
		return os.Getenv("DB_USERNAME_RO")
	}
}

func getDBPassword(choice int) string {
	switch choice {
	case 1:
		return os.Getenv("DB_PASSWORD_RW")
	default:
		return os.Getenv("DB_PASSWORD_RO")
	}
}

func getDBPort(choice int) string {
	switch choice {
	case 1, 2, 3:
		return os.Getenv("DB_PORT_MASTER")
	default:
		return os.Getenv("DB_PORT_SLAVE")
	}
}

func getAppPort(choice int) string {
	switch choice {
	case 1:
		return viper.GetString("app.port1")
	case 2:
		return viper.GetString("app.port2")
	case 3:
		return viper.GetString("app.port3")
	default:
		return viper.GetString("app.port1")
	}
}
