package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MainConnect() (*pgxpool.Pool, error) {
	host := os.Getenv("DB_MAIN_HOST")
	if host == "" {
		return nil, errors.New("environment variable DB_MAIN_HOST required")
	}

	user := os.Getenv("DB_MAIN_USER")
	if user == "" {
		return nil, errors.New("environment variable DB_MAIN_USER required")
	}

	password := os.Getenv("DB_MAIN_PASSWORD")
	if password == "" {
		return nil, errors.New("environment variable DB_MAIN_PASSWORD required")
	}

	port := os.Getenv("DB_MAIN_PORT")
	if port == "" {
		port = "5432"
	}

	database := os.Getenv("DB_MAIN_NAME")
	if database == "" {
		database = "main"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)

	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return nil, err
	}

	return conn, nil
}