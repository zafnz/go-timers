package main

import (
	"fmt"
	"os"
	"time"

	"github.com/zafnz/go-timers"
)

func main() {
	// Time how long we take
	defer fmt.Fprintf(os.Stderr, "---\nTimings:\n%s\n", timers.GlobalTimers)
	defer timers.New("main()").Start().Stop()

	// Time how long it takes to count to a hundred million
	timers.New("Count to a hundred million").Measure(func() {
		for i := 0; i < 100000000; i++ {
			// NOOP
		}
	})
	t := timers.New("Do things").Tag("Tag1").Tag("Tag2").Start()
	fmt.Println("Do some things...")
	t.Stop()
	OtherThings()
	/* Output is something like this:
	Do some things...
	---
	Timings:
	main(): 25.980ms
	Count to a hundred million: 25.948ms
	Do things: 0.009ms tags:(Tag1,Tag2)
	OtherThings(): 0.021ms
	*/
}

func OtherThings() {
	defer timers.New("OtherThings()").Start().Stop()
	time.Sleep(10 * time.Nanosecond)
}
