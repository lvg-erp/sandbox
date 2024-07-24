package main

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"os"
	"psqlprof/server/db"
	"psqlprof/server/internal/handler/command"
	command_cache "psqlprof/server/internal/repository/cache"
	command_repo "psqlprof/server/internal/repository/command"
	"psqlprof/server/internal/router"
	command_service "psqlprof/server/internal/service/command"
)

func main() {
	log.Printf("Init config....")
	if err := initConfig(); err != nil {
		log.Fatalf("error initializing configs: %s", err.Error())
	}

	log.Printf("Starting DB...")
	dbConn, err := db.NewDatabase(db.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	})

	if err != nil {
		log.Fatal("could initialized database connection: %s", err.Error())
	}
	defer dbConn.Close()

	if err = dbConn.Migrate(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("could not run migrations: %s", err.Error())
	}

	commandRep := command_repo.NewRepository(dbConn.GetDB())
	scriptsCache := command_cache.NewCache()
	execCmdCache := command_cache.NewCache()
	commandSvc := command_service.NewService(commandRep, scriptsCache, execCmdCache)
	defer func() { commandSvc.StopRunner() }()
	CommandHandler := command.NewHandler(commandSvc)

	r := router.InitRouter(CommandHandler)

	log.Printf("Starting server...")

	if err = router.Start(r, fmt.Sprintf("%s:%s", os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT"))); err != nil {
		log.Fatalf("could not start server: %s", err.Error())
	}
}

func initConfig() error {
	if err := godotenv.Load("././server/configs/.env"); err != nil {
		return fmt.Errorf("error loading .env file: %w ", err)
	}

	viper.AutomaticEnv()

	return nil
}
