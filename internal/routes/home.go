package routes

import (
	"log"
	"net/http"
	"text/template"

	"github.com/patoui/logme/internal/models"
)

func Home(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	// TODO get account ID from route param
	// accountId := r.Context().Value(accountIdKey).(int)

	logs, mapErr := models.List(dbLogs, 321, q)

	if mapErr != nil {
		log.Fatal(mapErr)
	}

	t, _ := template.ParseFiles(
		"templates/base.html",
		"templates/index.html",
	)

	data := struct {
		Title string
		Logs  []models.Log
	}{
		Title: "Home",
		Logs:  logs,
	}

	err := t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}
