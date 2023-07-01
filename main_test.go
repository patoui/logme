package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/patoui/logme/internal/models"
)

func setupTest() *Server {
	primaryIndex := "logs_test"
	s := Setup(map[string]string{
		"PRIMARY_INDEX": primaryIndex,
	})
	s.Db.DeleteIndex(primaryIndex)
	s.Db.Index(primaryIndex)
	return s
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
	s := setupTest()

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
	s := setupTest()

	br := strings.NewReader(`{"name":"error.log", "timestamp":"2022-12-31 12:36:58", "content":"this is a log entry", "account_id":321}`)
	req, _ := http.NewRequest("POST", "/log/321", br)
	req.Header.Add("Content-Type", "application/json")

	response := executeRequest(req, s)

	checkResponseCode(t, http.StatusCreated, response.Code)
	require.Equal(t, "{\"message\":\"Log successfully processed.\"}\n", response.Body.String())
}

func TestLogList(t *testing.T) {
	s := setupTest()

	br := strings.NewReader(`{"name":"error.log", "timestamp":"2022-12-31 12:36:58", "content":"this is a log entry", "account_id":321}`)
	req, _ := http.NewRequest("POST", "/log/321", br)
	req.Header.Add("Content-Type", "application/json")

	response := executeRequest(req, s)

	checkResponseCode(t, http.StatusCreated, response.Code)
	require.Equal(t, "{\"message\":\"Log successfully processed.\"}\n", response.Body.String())

	// TODO: find better approach
	// sleep to wait for document addition to occur
	time.Sleep(500 * time.Millisecond)

	gReq, _ := http.NewRequest("GET", "/log/321", nil)

	gResponse := executeRequest(gReq, s)

	checkResponseCode(t, http.StatusOK, gResponse.Code)
	checkResponseContentType(t, "application/json", gResponse.Header().Get("Content-Type"))
	var logs []models.Log
	json.NewDecoder(gResponse.Body).Decode(&logs)
	require.EqualValues(t, 1, len(logs))
	lastLog := logs[len(logs)-1]
	require.EqualValues(t, "this is a log entry", lastLog.Content)
	require.EqualValues(t, 321, lastLog.AccountId)
}
