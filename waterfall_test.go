package timers_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zafnz/go-timers"
)

func TestWaterfallHtml(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	handler := timers.WaterfallHandler()
	handler.ServeHTTP(response, request)

	result := response.Result()
	if result.StatusCode != 200 {
		t.Fatalf("Did not get back 200 for serving waterfall (%d)", result.StatusCode)
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "go-timers Waterfall Inspector") {
		t.Error("index.html did not contain go-timers waterfall header")
	}
}

func TestWaterfallJavascript(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/index.js", nil)
	handler := timers.WaterfallHandler()
	handler.ServeHTTP(response, request)

	result := response.Result()
	if result.StatusCode != 200 {
		t.Fatalf("Did not get back 200 for serving waterfall (%d)", result.StatusCode)
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Server-Timing") {
		t.Error("index.js did not contain Server-Timing ")
	}
}
