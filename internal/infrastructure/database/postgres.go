package database

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func DBConnect() *gorm.DB {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	schema := os.Getenv("DB_SCHEMA")
	connection := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo search_path=%s",
		host, user, password, dbname, port, schema,
	)
	db, err := gorm.Open(postgres.Open(connection), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Panic("Erro ao conectar com o db")
	}
	autoMigrate, err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE"))
	if err != nil {
		log.Panic("Erro ao identificar auto migrate")
	}

	if autoMigrate {
		db.AutoMigrate()
	}
	return db
}
