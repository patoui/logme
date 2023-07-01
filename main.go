package main

import (
	"log"
	"net/http"
	"os"

	chi "github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/meilisearch/meilisearch-go"
	"github.com/patoui/logme/internal/logme"
)

func main() {
	LoadEnv()
	s := CreateNewServer()
	s.MountHandlers()

	port := os.Getenv("APP_PORT")
	if len(port) == 0 {
		port = "8080"
	}

	http.ListenAndServe(":"+port, s.Router)
}

type Server struct {
	Router *chi.Mux
	Db     *meilisearch.Client
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func CreateNewServer() *Server {
	conn, err := logme.Connection()
	if err != nil {
		panic(err)
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