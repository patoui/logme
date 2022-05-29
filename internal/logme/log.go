package logme

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TODO: clean up this whole file

type key string

const (
	accountIdKey key = "accountId"
	logIdKey     key = "logId"
)

type createLog struct {
	AccountId uint32 `json:"account_id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

type Log struct {
	Uuid      *uuid.UUID `ch:"uuid"`
	Name      string     `ch:"name"`
	AccountId uint32     `ch:"account_id"`
	DateTime  time.Time  `ch:"dt"`
	Content   string     `ch:"content"`
}

type createValidationErr struct {
	Message string            `json:"timestamp"`
	Errors  map[string]string `json:"errors"`
}

var db driver.Conn

func Routes(r *chi.Mux, dbInstance driver.Conn) {
	db = dbInstance

	r.Route("/log", func(r chi.Router) {
		r.Get("/", List)
		r.Post("/", Create)

		r.Route("/{accountId:[0-9]+}/{logId:[a-zA-Z0-9\\-]+}", func(r chi.Router) {
			r.Use(LogContext)
			r.Get("/", Read)
		})
	})
}

func LogContext(next http.Handler) http.Handler {
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

		ctx = context.WithValue(ctx, logIdKey, logUuid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func List(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Made it to the list!"))
}

func Create(w http.ResponseWriter, r *http.Request) {
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

	if cl.AccountId == 0 || cl.Name == "" || cl.Timestamp == "" || cl.Content == "" {
		valErr.Message = "Validation error occurred."
		valErr.Errors = make(map[string]string)
	}

	if cl.AccountId == 0 {
		valErr.Errors["account_id"] = "'account_id' field is required."
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

	if err := createLogTable(); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	// TODO: protect against SQL injection
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = db.AsyncInsert(
		ctx,
		fmt.Sprintf(
			`INSERT INTO logs (account_id, dt, name, content) VALUES (%d, '%s', '%s', '%s')`,
			cl.AccountId,
			cl.Timestamp,
			cl.Name,
			cl.Content,
		),
		false,
	)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Log successfully processed.",
	})
}

func Read(w http.ResponseWriter, r *http.Request) {
	accountId := r.Context().Value(accountIdKey).(int)
	logId := r.Context().Value(logIdKey).(uuid.UUID)

	w.Header().Set("Content-Type", "application/json")

	if err := db.Ping(context.Background()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	sql := fmt.Sprintf("SELECT * FROM logs WHERE account_id = %d AND uuid = '%s'", accountId, logId.String())

	var currentLog Log
	if err := db.QueryRow(ctx, sql).ScanStruct(&currentLog); err != nil {
		// TODO: update to user friendly error message
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(currentLog)
}

func doesFileExist(filePath string) bool {
	_, filePathErr := os.Stat(filePath)
	return !os.IsNotExist(filePathErr)
}

// TODO: move to command to run DB migrations
func createLogTable() error {
	filePath := "internal/logme/migrations/000001_create_log_table.sql"

	if !doesFileExist(filePath) {
		return errors.New("migration file does not exists")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
		return errors.New("unable to read file: " + filePath)
	}

	err = db.Exec(context.Background(), string(content))

	if err != nil {
		return err
	}

	return nil
}
