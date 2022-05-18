package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	timers "github.com/zafnz/go-timers"
)

//go:embed fake-timings.json
var jsonData []byte

var fakeTimings *timers.TimerSet

func main() {
	// Grab a new TimerSet
	fakeTimings = timers.From(context.Background())
	err := json.Unmarshal(jsonData, fakeTimings)
	if err != nil {
		fmt.Printf("Failed to unmarshal timers json: %s\n", err.Error())
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api", fakeTimerData)
	mux.Handle("/waterfall/", http.StripPrefix("/waterfall/", timers.WaterfallHandler()))
	fmt.Println("Serving on http://127.0.0.1:3000")
	log.Fatal(http.ListenAndServe("127.0.0.1:3000", mux))
}

func fakeTimerData(w http.ResponseWriter, r *http.Request) {
	fakeTimings.AddHeader(w)
	w.WriteHeader(200)
	// Output body
	fmt.Fprintf(w, "See Server-Timing header, or view this page at /waterfall/")
}
