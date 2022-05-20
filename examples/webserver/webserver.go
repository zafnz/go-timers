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
	handler := timers.Middleware(mux, timers.MiddlewareOptions{})
	log.Fatal(http.ListenAndServe("127.0.0.1:3000", handler))
}

func apiSampleEndpoint(w http.ResponseWriter, r *http.Request) {
	// Note: This timer will not have it's duration in the header, as the header is sent before this function
	// exits (at w.WriteHeader and at Fprintf)
	defer timers.From(r.Context()).New("apiSample").Start().Stop()

	doWork()

	// Here we wrap apiFunc1 with it's own TimerSet, so all of it's timers it creates are all logically
	// grouped together.
	timers.From(r.Context()).Wrap(r.Context(), "calling apiFunc to do work", func(ctx context.Context) {
		apiFunc1(ctx)
	})
	// Where as apiFunc2, if it creates timers, it's timers will be in this parent context, not it's own
	// separate timer group.
	apiFunc2(r.Context())

	// When we write a header or body all timers will be snapshotted in place.
	w.WriteHeader(200)
	// Output body
	fmt.Fprintf(w, "Request ended")

}

func doWork() {
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
}

func apiFunc1(ctx context.Context) {
	defer timers.From(ctx).New("apiFunc1").Start().Stop()
	// Doing work
	doWork()

	t := timers.From(ctx).New("Downstream1").Start()
	doWork() // Make a call downstream
	t.Stop()
	apiFunc1a(ctx)
	apiFunc1b(ctx)

}
func apiFunc1a(ctx context.Context) {
	defer timers.From(ctx).New("apiFunc1a").Start().Stop()
	// Doing work
	doWork()

	t := timers.From(ctx).New("Downstream2").Start()
	doWork() // Make a call downstream
	t.Stop()
}
func apiFunc1b(ctx context.Context) {
	defer timers.From(ctx).New("apiFunc1b").Start().Stop()
	// Doing work
	doWork()

	t := timers.From(ctx).New("Downstream3").Start()
	doWork() // Make a call downstream
	t.Stop()
}
func apiFunc2(ctx context.Context) {
	defer timers.From(ctx).New("apiFunc2").Start().Stop()
	doWork()
	apiFunc3(ctx)
	// Lots of work
	doWork()
	doWork()
	doWork()
	apiFunc4(ctx)
}
func apiFunc3(ctx context.Context) {
	defer timers.From(ctx).New("apiFunc3").Start().Stop()
	doWork()
}
func apiFunc4(ctx context.Context) {
	defer timers.From(ctx).New("apiFunc4").Start().Stop()
	doWork()
}
