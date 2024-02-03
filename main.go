package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/meilisearch/meilisearch-go"

	"github.com/patoui/logme/internal/db"
	"github.com/patoui/logme/internal/routes"
)

func main() {
	s := Setup(map[string]string{})
	port := os.Getenv("APP_PORT")
	if len(port) == 0 {
		port = "8080"
	}

	http.ListenAndServe(":"+port, s.Router)
}

func Setup(overrides map[string]string) *Server {
	LoadEnv()
	for k, v := range overrides {
		os.Setenv(k, v)
	}
	s := CreateNewServer()
	s.MountHandlers()
	return s
}

type Server struct {
	Router *chi.Mux
	Db     *meilisearch.Client
	Main   *pgxpool.Pool
	Logs   driver.Conn
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func CreateNewServer() *Server {
	conn, err := db.Connection()
	if err != nil {
		panic(err)
	}

	logsConn, err := db.LogsConnect()
	if err != nil {
		panic(err)
	}

	mainConn, err := db.MainConnect()
	if err != nil {
		panic(err)
	}

	return &Server{
		Router: chi.NewRouter(),
		Db:     conn,
		Main:   mainConn,
		Logs:   logsConn,
	}
}

func (s *Server) MountHandlers() {
	routes.RegisterRoutes(s.Router, s.Db, s.Logs, s.Main)
	s.Router.Get("/", routes.Home)
}
