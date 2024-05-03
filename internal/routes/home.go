package routes

import (
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/patoui/logme/internal/models"
)

func Home(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	// TODO get account ID from route param
	// accountId := r.Context().Value(accountIdKey).(int)
	var err error

	logs, err := models.List(dbLogs, 321, q)

	if err != nil {
		log.Fatal(err)
	}

	var t *template.Template
	t = template.Must(template.New("base.html").Funcs(template.FuncMap{
		"fdate": func(dt time.Time) string {
			return dt.Format(models.DateFormat)
		},
	}).ParseFiles(
		"templates/base.html",
		"templates/index.html",
	))

	data := struct {
		Title string
		Logs  []models.Log
	}{
		Title: "Home",
		Logs:  logs,
	}

	err = t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}
