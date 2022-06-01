package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

// checkResponseCode is a simple utility to check the response code
// of the response
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// checkResponseContentType is a simple utility to check
// "Content-Type" of the response
func checkResponseContentType(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Errorf("Expected content-type %s. Got %s\n", expected, actual)
	}
}

func TestHome(t *testing.T) {
	// Create a New Server Struct
	s := CreateNewServer()
	// Mount Handlers
	s.MountHandlers()

	// Create a New Request
	req, _ := http.NewRequest("GET", "/", nil)

	// Execute Request
	response := executeRequest(req, s)

	// Check the response code
	checkResponseCode(t, http.StatusOK, response.Code)

	// We can use testify/require to assert values, as it is more convenient
	require.Equal(t, "{\"message\":\"Welcome to LogMe!\"}\n", response.Body.String())
}

func TestLogCreate(t *testing.T) {
	LoadEnv()
	os.Setenv("DB_NAME", "logme_test")
	s := CreateNewServer()
	s.MountHandlers()

	br := strings.NewReader(`{"name":"error.log","timestamp":"2022-01-01 01:01:01", "content":"foobar", "account_id": 321}`)
	req, _ := http.NewRequest("POST", "/log/321", br)
	req.Header.Add("Content-Type", "application/json")

	response := executeRequest(req, s)

	checkResponseCode(t, http.StatusCreated, response.Code)
	require.Equal(t, "{\"message\":\"Log successfully processed.\"}\n", response.Body.String())
}

func TestLogList(t *testing.T) {
	LoadEnv()
	os.Setenv("DB_NAME", "logme_test")
	s := CreateNewServer()
	s.MountHandlers()

	br := strings.NewReader(`{"name":"error.log","timestamp":"2022-01-01 01:01:01", "content":"foobar"}`)
	req, _ := http.NewRequest("POST", "/log/321", br)
	req.Header.Add("Content-Type", "application/json")

	response := executeRequest(req, s)

	checkResponseCode(t, http.StatusCreated, response.Code)
	require.Equal(t, "{\"message\":\"Log successfully processed.\"}\n", response.Body.String())

	gReq, _ := http.NewRequest("GET", "/log/321", nil)

	gResponse := executeRequest(gReq, s)

	checkResponseCode(t, http.StatusOK, gResponse.Code)
	checkResponseContentType(t, "application/json", gResponse.Header().Get("Content-Type"))
	type log struct {
		Uuid      *uuid.UUID `json:"uuid"`
		Name      string     `json:"name"`
		AccountId uint32     `json:"account_id"`
		DateTime  string     `json:"dt"`
		Content   string     `json:"content"`
	}
	var logs []log
	json.NewDecoder(gResponse.Body).Decode(&logs)
	lastLog := logs[len(logs)-1]
	require.EqualValues(t, "foobar", lastLog.Content)
	require.EqualValues(t, 321, lastLog.AccountId)
}
