package logme

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/mitchellh/mapstructure"
)

// TODO: clean up this whole file

type key string

const (
	accountIdKey key = "accountId"
	logIdKey     key = "logId"
)

type Log struct {
	Uuid      *uuid.UUID `mapstructure:"uuid" json:"uuid"`
	Name      string     `mapstructure:"name" json:"name"`
	AccountId uint32     `mapstructure:"account_id" json:"account_id"`
	DateTime  time.Time  `mapstructure:"dt" json:"dt"`
	Content   string     `mapstructure:"content" json:"content"`
}

type OutputLog struct {
	Uuid      string `mapstructure:"uuid" json:"uuid"`
	Name      string `mapstructure:"name" json:"name"`
	AccountId uint32 `mapstructure:"account_id" json:"account_id"`
	DateTime  string `mapstructure:"timestamp" json:"timestamp"`
	Content   string `mapstructure:"content" json:"content"`
}

type createLog struct {
	AccountId uint32 `json:"account_id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

type createValidationErr struct {
	Message string            `json:"timestamp"`
	Errors  map[string]string `json:"errors"`
}

var db *meilisearch.Client

func RegisterRoutes(r *chi.Mux, dbInstance *meilisearch.Client) {
	db = dbInstance

	r.Route("/log/{accountId:[0-9]+}", func(r chi.Router) {
		r.Use(AccountContext)
		r.Get("/", List)
		r.Post("/", Create)
		// r.Route("/{logId:[a-zA-Z0-9\\-]+}", func(r chi.Router) {
		// 	r.Use(LogContext)
		// 	r.Get("/", Read)
		// })
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

func LogContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logId := chi.URLParam(r, string(logIdKey))
		if logId == "" {
			http.NotFound(w, r)
			return
		}

		logUuid, err := uuid.Parse(logId)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), logIdKey, logUuid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func List(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	q := r.URL.Query().Get("q")
	index := db.Index("logs")

	_, err := index.UpdateFilterableAttributes(&[]string{"account_id"})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	accountId := r.Context().Value(accountIdKey).(int)

	resp, err := index.Search(q, &meilisearch.SearchRequest{
        Filter: fmt.Sprintf("account_id = %d", accountId),
    })

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var logs []OutputLog
	errtwo := mapstructure.Decode(resp.Hits, &logs)
	if errtwo != nil {
		fmt.Println("Error decoding map to log struct:", errtwo)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func Create(w http.ResponseWriter, r *http.Request) {
	accountId := r.Context().Value(accountIdKey).(int)
	var unmarshalErr *json.UnmarshalTypeError
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("Content-Type is not application/json"))
		return
	}

	var cl createLog
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	err := d.Decode(&cl)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		var msg string
		w.WriteHeader(http.StatusUnprocessableEntity)
		if errors.As(err, &unmarshalErr) {
			msg = "Bad Request. Wrong Type provided for field " + unmarshalErr.Field
		} else {
			msg = "Bad Request " + err.Error()
		}
		json.NewEncoder(w).Encode(map[string]string{
			"message": msg,
		})
		return
	}

	var valErr createValidationErr

	if cl.Name == "" || cl.Timestamp == "" || cl.Content == "" {
		valErr.Message = "Validation error occurred."
		valErr.Errors = make(map[string]string)
	}

	if cl.Name == "" {
		valErr.Errors["name"] = "'name' field is required."
	}

	// TODO: validate timestamp format
	if cl.Timestamp == "" {
		valErr.Errors["timestamp"] = "'timestamp' field is required."
	}

	if cl.Content == "" {
		valErr.Errors["content"] = "'content' field is required."
	}

	if len(valErr.Message) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(valErr)
		return
	}

	index := db.Index("logs")

    id := uuid.New()
    documents := []map[string]interface{}{
    	{
			"uuid": id.String(),
			"account_id": accountId,
			"name": cl.Name,
			"timestamp": cl.Timestamp,
			"content": cl.Content,
		},
	}

	task, err := index.AddDocuments(documents, "uuid")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(task.TaskUID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Log successfully processed.",
	})
}

func doesFileExist(filePath string) bool {
	_, filePathErr := os.Stat(filePath)
	return !os.IsNotExist(filePathErr)
}
