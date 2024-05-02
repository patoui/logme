package routes

import (
	"context"
	"encoding/json"
	"fmt"
	syslog "log"
	"net/http"
	"strconv"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/patoui/logme/internal/models"
)

var dbLogs driver.Conn
var dbMain *pgxpool.Pool

const accountIdKey = "accountId"
const layout = "2006-01-02 15:04:05"

type createValidationErr struct {
	Message string            `json:"timestamp"`
	Errors  map[string]string `json:"errors"`
}

func RegisterRoutes(r *chi.Mux, logsConn driver.Conn, mainConn *pgxpool.Pool) {
	dbLogs = logsConn
	dbMain = mainConn

	r.Route("/log/{accountId:[0-9]+}", func(r chi.Router) {
		r.Use(AccountContext)
		r.Get("/", list)
		r.Post("/", Create)
	})
}

func AccountContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountId := chi.URLParam(r, string(accountIdKey))
		if accountId == "" {
			http.NotFound(w, r)
			return
		}

		accountIdInt, err := strconv.Atoi(accountId)
		if err != nil || accountIdInt == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": err.Error(),
			})
		}

		ctx := context.WithValue(r.Context(), accountIdKey, accountIdInt)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	accountId := r.Context().Value(accountIdKey).(int)

	logs, mapErr := models.List(dbLogs, accountId, q)
	if mapErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func Create(w http.ResponseWriter, r *http.Request) {
	accountId := r.Context().Value(accountIdKey).(int)
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("Content-Type is not application/json"))
		return
	}

	var log models.Log
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	decodeErr := d.Decode(&log)

	w.Header().Set("Content-Type", "application/json")

	if decodeErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("{'message': '%s'}", decodeErr.Error())))
		return
	}

	valErr := createValidationErr{
		Message: "",
		Errors:  make(map[string]string),
	}

	if log.Name == "" || !log.DateTime.IsSet() || log.Content == "" {
		valErr.Message = "Validation error occurred."
		valErr.Errors = make(map[string]string)
	}

	if log.Name == "" {
		valErr.Errors["name"] = "'name' field is required."
	}

	if log.DateTime.IsSet() {
		valErr.Errors["timestamp"] = "'timestamp' field is required."
	}

	if log.Content == "" {
		valErr.Errors["content"] = "'content' field is required."
	}

	if len(valErr.Message) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(valErr)
		return
	}

	log.AccountId = uint32(accountId)
	log.Uuid = uuid.New()

	docErr := log.Create(dbLogs)
	if docErr != nil {
		syslog.Println(docErr)
		json.NewEncoder(w).Encode(docErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Log successfully processed.",
	})
}
