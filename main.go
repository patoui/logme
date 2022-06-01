package main

import (
	"log"
	"net/http"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	chi "github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/patoui/logme/internal/logme"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	s := CreateNewServer()
	s.MountHandlers()
	http.ListenAndServe(":8080", s.Router)
}

type Server struct {
	Router *chi.Mux
	Db     driver.Conn
}

func CreateNewServer() *Server {
	conn, err := logme.Connection()
	if err != nil {
		panic("Unable to connect to the database.")
	}
	return &Server{
		Router: chi.NewRouter(),
		Db:     conn,
	}
}

func (s *Server) MountHandlers() {
	logme.RegisterRoutes(s.Router, s.Db)
	s.Router.Get("/", logme.Home)
}
