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

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/controller"
	dbpostgres "git.iu7.bmstu.ru/vai20u117/testing/src/internal/db/postgres"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"

	// _ "git.iu7.bmstu.ru/vai20u117/testing/src/swagger"
	muxhandlers "github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const requestsTimeout = 15 * time.Second

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetLogLoggerLevel(slog.LevelInfo)

	mustLoadConfigs()

	database := mustLoadDB(ctx)
	defer database.GetPool(ctx).Close()

	controller := controller.NewController(
		initPosterHandler(database),
		initListHandler(database),
		initListPosterHandler(database),
		initPosterRecordHandler(database),
		initAuthHandler(database, os.Getenv("ADMIN_SECRET")),
	)

	serverPort := ":" + "9000" // viper.GetString("port")
	router := controller.CreateRouter()
	http.Handle("/", router)

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

	slog.Info("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("Server shutting down")
}

func initPosterHandler(database *dbpostgres.Database) *controller.PosterHandler {
	posterRepository := repository.NewPosterRepository(database)
	posterService := service.NewPosterService(posterRepository)
	return controller.NewPosterHandler(posterService)
}

func initListHandler(database *dbpostgres.Database) *controller.ListHandler {
	listRepository := repository.NewListRepository(database)
	listService := service.NewListService(listRepository)
	return controller.NewListHandler(listService)
}

func initListPosterHandler(database *dbpostgres.Database) *controller.ListPosterHandler {
	listPosterRepository := repository.NewListPosterRepository(database)
	listPosterService := service.NewListPosterService(listPosterRepository)
	return controller.NewListPosterHandler(listPosterService)
}

func initPosterRecordHandler(database *dbpostgres.Database) *controller.PosterRecordHandler {
	historyRepository := repository.NewPosterRecordRepository(database)
	historyService := service.NewPosterRecordService(historyRepository)
	return controller.NewPosterRecordHandler(historyService)
}

func initAuthHandler(database *dbpostgres.Database, adminToken string) *controller.AuthHandler {
	userRepository := repository.NewUserRepository(database)
	authService := service.NewAuthService(userRepository, adminToken)
	return controller.NewAuthHandler(authService)
}

func mustLoadConfigs() {
	// if err := initConfig(); err != nil {
	// 	log.Fatal("Failed to init configs: ", err)
	// }
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load env variables: ", err)
	}
}

func mustLoadDB(ctx context.Context) *dbpostgres.Database {
	database, err := dbpostgres.NewDB(ctx, &dbpostgres.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
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
