//go:build integration
// +build integration

package tests

import (
	"log"

	testDB "git.iu7.bmstu.ru/vai20u117/testing/src/internal/tests/postgres"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	db *testDB.TDB
)

func init() {
	if err := initConfig(); err != nil {
		log.Fatal("Failed to init configs: ", err)
	}
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load env variables: ", err)
	}

	db = testDB.NewFromEnv()
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
