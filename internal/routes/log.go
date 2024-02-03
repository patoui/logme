package routes

import (
	"context"
	"encoding/json"
	syslog "log"
	"net/http"
	"strconv"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meilisearch/meilisearch-go"

	"github.com/patoui/logme/internal/models"
)

var dbInstance *meilisearch.Client
var dbLogs driver.Conn
var dbMain *pgxpool.Pool

const accountIdKey = "accountId"
const layout = "2006-01-02 15:04:05"

type createValidationErr struct {
    Message string            `json:"timestamp"`
    Errors  map[string]string `json:"errors"`
}

func RegisterRoutes(r *chi.Mux, dbGlobal *meilisearch.Client, logsConn driver.Conn, mainConn *pgxpool.Pool) {
    dbInstance = dbGlobal
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

    logs, mapErr := models.List(dbInstance, accountId, q)
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

    var cl models.CreateLog
    d := json.NewDecoder(r.Body)
    d.DisallowUnknownFields()
    d.Decode(&cl)

    w.Header().Set("Content-Type", "application/json")

    var valErr createValidationErr

    if cl.Name == "" || !cl.Timestamp.IsSet() || cl.Content == "" {
        valErr.Message = "Validation error occurred."
        valErr.Errors = make(map[string]string)
    }

    if cl.Name == "" {
        valErr.Errors["name"] = "'name' field is required."
    }

    if !cl.Timestamp.IsSet() {
        valErr.Errors["timestamp"] = "'timestamp' field is required."
    }

    if cl.Content == "" {
        valErr.Errors["content"] = "'content' field is required."
    }

    if len(valErr.Message) > 0 {
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(valErr)
        return
    }

    cl.AccountId = accountId

    docErr := cl.Create(dbInstance, dbLogs)
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
