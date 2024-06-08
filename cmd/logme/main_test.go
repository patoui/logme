package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patoui/logme/internal/model"
)

func setupTest() (*Server, func()) {
	s := Setup(map[string]string{
		"DB_LOGS_DEBUG":      "false",
		"DB_LOGS_NAME":       "logs_test",
		"DB_LOGS_ASYNC_WAIT": "true",
	})

	teardown := func() {
		s.Logs.Exec(context.Background(), "TRUNCATE TABLE IF EXISTS logs_test.logs")
		s.Logs.Close()
	}

	return s, teardown
}

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
	s, teardown := setupTest()
	defer teardown()

	// Create a New Request
	req, _ := http.NewRequest("GET", "/", nil)

	// Execute Request
	response := executeRequest(req, s)

	// Check the response code
	checkResponseCode(t, http.StatusOK, response.Code)

	// We can use testify/require to assert values, as it is more convenient
	assert.Equal(t, "{\"message\":\"Welcome to LogMe!\"}\n", response.Body.String())
}

func TestLogCreate(t *testing.T) {
	s, teardown := setupTest()
	defer teardown()

	br := strings.NewReader(`{"name":"error.log", "timestamp":"2022-12-31 12:36:58", "content":"this is a log entry", "account_id":321}`)
	req, _ := http.NewRequest("POST", "/log/321", br)
	req.Header.Add("Content-Type", "application/json")

	response := executeRequest(req, s)

	checkResponseCode(t, http.StatusCreated, response.Code)
	assert.Equal(t, "{\"message\":\"Log successfully processed.\"}\n", response.Body.String())
}

func TestLogList(t *testing.T) {
	s, teardown := setupTest()
	defer teardown()

	br := strings.NewReader(`{"name":"error.log", "timestamp":"2022-12-31 12:36:58", "content":"this is a log entry", "account_id":321}`)
	req, _ := http.NewRequest("POST", "/log/321", br)
	req.Header.Add("Content-Type", "application/json")

	response := executeRequest(req, s)

	checkResponseCode(t, http.StatusCreated, response.Code)
	assert.Equal(t, "{\"message\":\"Log successfully processed.\"}\n", response.Body.String())

	gReq, _ := http.NewRequest("GET", "/log/321", nil)
	gResponse := executeRequest(gReq, s)

	checkResponseCode(t, http.StatusOK, gResponse.Code)
	checkResponseContentType(t, "application/json", gResponse.Header().Get("Content-Type"))
	var logs []model.Log
	json.NewDecoder(gResponse.Body).Decode(&logs)
	assert.Equal(t, 1, len(logs))
	lastLog := logs[len(logs)-1]
	assert.Equal(t, "this is a log entry", lastLog.Content)
	assert.Equal(t, uint32(321), *lastLog.AccountId)
}
