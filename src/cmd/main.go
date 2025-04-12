package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	database := mustLoadDB(ctx)
	defer database.GetPool(ctx).Close()

	controllerV2 := createControllerV2(database)

	serverPort := ":" + getAppPort()
	router := mux.NewRouter()
	// controller.CreateRouter(router)
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

func createControllerV2(database *dbpostgres.Database) *apiv2.Controller {
	return apiv2.NewController(
		initPosterHandlerV2(database),
		initListHandlerV2(database),
		initListPosterHandlerV2(database),
		initPosterRecordHandlerV2(database),
		initAuthHandlerV2(database, os.Getenv(adminSecretEnv)),
	)
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

func mustLoadDB(ctx context.Context) *dbpostgres.Database {
	database, err := dbpostgres.NewDB(ctx, &dbpostgres.DBConfig{
		Host: viper.GetString("db.host"),
		// Port:     viper.GetString("db.port"),
		Port:     getDBPort(),
		Username: getDBUsername(),
		Password: getDBPassword(),
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

func getDBUsername() string {
	return os.Getenv("DB_USERNAME")
}

func getDBPassword() string {
	return os.Getenv("DB_PASSWORD")
}

func getDBPort() string {
	return os.Getenv("DB_PORT")
}

func getAppPort() string {
	return viper.GetString("app.port")
}
