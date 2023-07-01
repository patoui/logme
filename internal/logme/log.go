package logme

import (
	"context"
	"encoding/json"
	"fmt"
	syslog "log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/mitchellh/mapstructure"
)

var db *meilisearch.Client

const accountIdKey = "accountId"
const layout = "2006-01-02 15:04:05"

type Log struct {
	Uuid      *uuid.UUID `mapstructure:"uuid" json:"uuid"`
	Name      string     `mapstructure:"name" json:"name"`
	AccountId uint32     `mapstructure:"account_id" json:"account_id"`
	DateTime  CustomTime `mapstructure:"timestamp" json:"timestamp"`
	Content   string     `mapstructure:"content" json:"content"`
}

type createLog struct {
	AccountId uint32     `json:"account_id"`
	Name      string     `json:"name"`
	Timestamp CustomTime `json:"timestamp"`
	Content   string     `json:"content"`
}

type createValidationErr struct {
	Message string            `json:"timestamp"`
	Errors  map[string]string `json:"errors"`
}

// Move custom time to separate dir
type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(layout, s)
	return
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(layout))), nil
}

var nilTime = (time.Time{}).UnixNano()

func (ct *CustomTime) IsSet() bool {
	return ct.UnixNano() != nilTime
}

func decode(input, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			stringToUUIDHookFunc(),
			stringToCustomTimeHookFunc(),
		),
		Result: &output,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func stringToUUIDHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(uuid.UUID{}) {
			return data, nil
		}

		return uuid.Parse(data.(string))
	}
}

func stringToCustomTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(CustomTime{}) {
			return data, nil
		}

		tm, _ := time.Parse(layout, data.(string))

		return CustomTime{tm}, nil
	}
}

func RegisterRoutes(r *chi.Mux, dbInstance *meilisearch.Client) {
	db = dbInstance

	r.Route("/log/{accountId:[0-9]+}", func(r chi.Router) {
		r.Use(AccountContext)
		r.Get("/", List)
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

	var logs []Log
	mapErr := decode(resp.Hits, &logs)
	// mapErr := mapstructure.Decode(resp.Hits, &logs)
	if mapErr != nil {
		fmt.Println("Error decoding map to logs struct:", mapErr)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapErr)
		return
	}

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

	var cl createLog
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

	index := db.Index("logs")

	id := uuid.New()
	documents := []map[string]interface{}{
		{
			"uuid":       id.String(),
			"account_id": accountId,
			"name":       cl.Name,
			"timestamp":  cl.Timestamp,
			"content":    cl.Content,
		},
	}

	_, docErr := index.AddDocuments(documents, "uuid")
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
