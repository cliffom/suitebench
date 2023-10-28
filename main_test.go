package main

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type MockRequester struct {
	Response *http.Response
	Error    error
}

func (m *MockRequester) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(""))}, m.Error
}

func TestApp_Run(t *testing.T) {
	app := &App{
		URL:         "http://example.com",
		NumRequests: 10,
		Concurrency: 5,
		Requester:   &MockRequester{}, // No need to set Response here unless you need a specific behavior
		Out:         &bytes.Buffer{},
	}
	app.Run()

	// Now you can check app.Out, which is a bytes.Buffer, to see if the output was what you expected.
	output := app.Out.(*bytes.Buffer).String()
	expectedOutputPart := "Finished 10 requests"
	if !bytes.Contains([]byte(output), []byte(expectedOutputPart)) {
		t.Errorf("Expected output to contain %q, but got %q", expectedOutputPart, output)
	}

	// Add more checks as necessary.
}
