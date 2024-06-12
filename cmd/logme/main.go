package main

import (
	"net/http"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rueian/valkey-go"

	"github.com/patoui/logme/internal/db"
	"github.com/patoui/logme/internal/helper"
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
	helper.LoadEnv()
	for k, v := range overrides {
		os.Setenv(k, v)
	}
	s := CreateNewServer()
	s.MountHandlers()
	return s
}

type Server struct {
	Router *chi.Mux
	Main   *pgxpool.Pool
	Logs   driver.Conn
	Cache  valkey.Client
}

func CreateNewServer() *Server {
	logsConn, err := db.LogsConnect()
	if err != nil {
		panic(err)
	}

	mainConn, err := db.MainConnect()
	if err != nil {
		panic(err)
	}

	cacheClient, err := db.Cache()
	if err != nil {
		panic(err)
	}

	return &Server{
		Router: chi.NewRouter(),
		Main:   mainConn,
		Logs:   logsConn,
		Cache:  cacheClient,
	}
}

func (s *Server) MountHandlers() {
	routes.RegisterRoutes(s.Router, s.Logs, s.Main)
	s.Router.Get("/", routes.Home)
	s.Router.Get("/ws", routes.Websocket)
}
