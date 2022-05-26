package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		t.Errorf("Expected response code %s. Got %s\n", expected, actual)
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
	s := CreateNewServer()
	s.MountHandlers()

	br := strings.NewReader(`{"name":"error.log","timestamp":"2022-01-01 01:01:01", "content":"foobar"}`)
	req, _ := http.NewRequest("POST", "/log", br)
	req.Header.Add("Content-Type", "application/json")

	response := executeRequest(req, s)

	checkResponseCode(t, http.StatusOK, response.Code)
	require.Equal(t, "{\"message\":\"Log successfully processed.\"}\n", response.Body.String())
}

func TestLogRead(t *testing.T) {
	s := CreateNewServer()
	s.MountHandlers()

	br := strings.NewReader(`{"name":"error.log","timestamp":"2022-01-01 01:01:01", "content":"foobar"}`)
	req, _ := http.NewRequest("POST", "/log", br)
	req.Header.Add("Content-Type", "application/json")

	response := executeRequest(req, s)

	checkResponseCode(t, http.StatusOK, response.Code)
	require.Equal(t, "{\"message\":\"Log successfully processed.\"}\n", response.Body.String())

	gReq, _ := http.NewRequest("GET", "/log/error.log", nil)

	gResponse := executeRequest(gReq, s)

	checkResponseCode(t, http.StatusOK, gResponse.Code)
	checkResponseContentType(t, gResponse.Result().Header.Get("Content-Type"), "text/plain")
	// verify last entry is what we previously added
	require.True(t, strings.HasSuffix(gResponse.Body.String(), "foobar\n"))
}
