package timers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zafnz/go-timers"
)

func ExampleMiddleware() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		defer timers.From(r.Context()).New("Api Function").Start().Stop()
	})
	// Add the waterfall handler too.
	// Note the trailing slashes are important!
	mux.Handle("/waterfall/", http.StripPrefix("/waterfall/", timers.WaterfallHandler()))
	handler := timers.Middleware(mux, timers.MiddlewareOptions{})
	http.ListenAndServe("127.0.0.1:3000", handler)
}

func TestMiddleware(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timers.From(r.Context()).New("test").Start().Stop()
	})

	middleware := timers.Middleware(handler, timers.MiddlewareOptions{})
	middleware.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatal("Something broke")
	}
	timingHeader := rr.Header().Get("Server-Timing")
	if timingHeader == "" {
		t.Fatal("No server timing header")
	}
	if !strings.Contains(timingHeader, "descr=\"test\"") {
		t.Error("Server-Timing does not contain test timer")
	}
}

func TestMiddlewareWrite(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timers.From(r.Context()).New("test").Start().Stop()
		fmt.Fprintln(w, "output")
	})

	middleware := timers.Middleware(handler, timers.MiddlewareOptions{})
	middleware.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatal("Something broke")
	}
	timingHeader := rr.Header().Get("Server-Timing")
	if timingHeader == "" {
		t.Fatal("No server timing header")
	}
	if !strings.Contains(timingHeader, "descr=\"test\"") {
		t.Error("Server-Timing does not contain test timer")
	}
}

func TestMiddlewareWriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timers.From(r.Context()).New("test").Start().Stop()
		w.WriteHeader(300)
	})

	middleware := timers.Middleware(handler, timers.MiddlewareOptions{})
	middleware.ServeHTTP(rr, req)
	if rr.Code != 300 {
		t.Fatal("Something broke")
	}
	timingHeader := rr.Header().Get("Server-Timing")
	if timingHeader == "" {
		t.Fatal("No server timing header")
	}
	if !strings.Contains(timingHeader, "descr=\"test\"") {
		t.Error("Server-Timing does not contain test timer")
	}
}
