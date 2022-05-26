package logme

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

const storagePath = "log_storage/"

type createLog struct {
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

type createValidationErr struct {
	Message string            `json:"timestamp"`
	Errors  map[string]string `json:"errors"`
}

func Create(w http.ResponseWriter, r *http.Request) {
	// TODO: clean up this whole file

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

	_, repositoryDirErr := os.Stat(storagePath)

	// check if log storage directory exists
	if os.IsNotExist(repositoryDirErr) {
		// create log storage directory
		if err := os.MkdirAll(storagePath, 0700); err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": err.Error(),
			})
			return
		}
	}

	filename := "log_storage/" + cl.Name

	// check if file exists
	if !doesFileExist(filename) {
		// create file
		_, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": err.Error(),
			})
			return
		}
	}

	// create file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": err.Error(),
		})
		return
	}

	defer f.Close()

	// TODO: insert to clickhouse
	// write to the file
	if _, err = f.WriteString(cl.Content + "\n"); err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Log successfully processed.",
	})
}

func Read(w http.ResponseWriter, r *http.Request) {
	// TODO: add query param
	logId := chi.URLParam(r, "id")

	filePath := storagePath + "/" + logId

	if !doesFileExist(filePath) {
		http.NotFound(w, r)
		return
	}

	// TODO: stream contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

func doesFileExist(filePath string) bool {
	_, filePathErr := os.Stat(filePath)
	return !os.IsNotExist(filePathErr)
}
