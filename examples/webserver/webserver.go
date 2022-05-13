package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/zafnz/go-timers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", apiSampleEndpoint)
	mux.Handle("/waterfall/", http.StripPrefix("/waterfall/", timers.WaterfallHandler()))
	log.Fatal(http.ListenAndServe("127.0.0.1:3000", timerMiddleware(mux)))
}

func doWork() {
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
}

func apiFunc1(ctx context.Context) {
	defer timers.Get(ctx).New("apiFunc1").Start().Stop()
	// Doing work
	doWork()

	t := timers.Get(ctx).New("Downstream1").Start()
	doWork() // Make a call downstream
	t.Stop()
	apiFunc1a(ctx)
	apiFunc1b(ctx)

}
func apiFunc1a(ctx context.Context) {
	defer timers.Get(ctx).New("apiFunc1a").Start().Stop()
	// Doing work
	doWork()

	t := timers.Get(ctx).New("Downstream2").Start()
	doWork() // Make a call downstream
	t.Stop()
}
func apiFunc1b(ctx context.Context) {
	defer timers.Get(ctx).New("apiFunc1b").Start().Stop()
	// Doing work
	doWork()

	t := timers.Get(ctx).New("Downstream3").Start()
	doWork() // Make a call downstream
	t.Stop()
}
func apiFunc2(ctx context.Context) {
	defer timers.Get(ctx).New("apiFunc2").Start().Stop()
	doWork()
	apiFunc3(ctx)
	// Lots of work
	doWork()
	doWork()
	doWork()
	apiFunc4(ctx)
}
func apiFunc3(ctx context.Context) {
	defer timers.Get(ctx).New("apiFunc3").Start().Stop()
	doWork()
}
func apiFunc4(ctx context.Context) {
	defer timers.Get(ctx).New("apiFunc4").Start().Stop()
	doWork()
}

func apiSampleEndpoint(w http.ResponseWriter, r *http.Request) {

	doWork()
	timers.Get(r.Context()).Wrap(r.Context(), "calling apiFunc to do work", func(ctx context.Context) {
		apiFunc1(ctx)
	})
	apiFunc2(r.Context())
	// Output headers
	timers.Get(r.Context()).AddHeader(w)
	w.WriteHeader(200)
	// Output body
	fmt.Fprintf(w, "Request ended")

}

func timerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = timers.NewContext(ctx)
		name := r.URL.Path

		timers.Get(ctx).New(name).Start()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
