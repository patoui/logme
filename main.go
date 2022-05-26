package main

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/patoui/logme/internal/logme"
)

func main() {
	s := CreateNewServer()
	s.MountHandlers()
	http.ListenAndServe(":8080", s.Router)
}

type Server struct {
	Router *chi.Mux
	// Db, config can be added here
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers() {
	s.Router.Get("/", logme.Home)
	s.Router.Post("/log", logme.Create)
	s.Router.Get("/log/{id:^[a-zA-Z0-9\\-\\.]+}", logme.Read)
}
