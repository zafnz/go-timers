package timers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/zafnz/go-timers"
)

func majorWork(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Println("Some other major work starting")
	time.Sleep(time.Duration(rand.Intn(1000) * int(time.Millisecond)))
	fmt.Println("Some other major work finished")
	wg.Done()

}
func moreMajorWork(ctx context.Context) {
	fmt.Println("more major work starting")
	time.Sleep(time.Duration(rand.Intn(1000) * int(time.Millisecond)))
	timers.From(ctx).Wrap(ctx, "subwork", func(ctx context.Context) {
		fmt.Println("Subwork going on")
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		fmt.Println("subwork finished")
	})
	fmt.Println("more major work finished")
}

func Example() {
	// Create a new TimerSet in a new context.
	ctx := timers.NewContext(context.Background())

	// do things
	func() {
		defer timers.From(ctx).New("Measure Sleep").Start().Stop()
		time.Sleep(100 * time.Millisecond)
	}()
	t := timers.From(ctx).New("Count to a hundred million")
	t.Start()
	for i := 0; i < 100000000; i++ {
	}
	t.Stop()

	// Create a new context and do a bunch of major work under
	// that context, bundling all those timers in their own collection.
	// All the timers are retrievable from the parent context.
	func() {
		ctx, t := timers.NewContextWithTimer(ctx, "All major work")
		t.Start()
		defer t.Stop()

		wg := sync.WaitGroup{}
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go timers.From(ctx).Wrap(ctx, "majorWork", func(ctx context.Context) {
				majorWork(ctx, &wg)
			})
		}
		wg.Wait()
		fmt.Println("All major work complete")
	}()

	workCtx, t := timers.NewContextWithTimer(ctx, "More work")
	t.Start()
	moreMajorWork(workCtx)
	t.Stop()

	fmt.Println("\nLets see how we did:")
	for _, t := range timers.From(ctx).AllDeep() {
		fmt.Println(t)
	}
	fmt.Println("\n\nAs a tree")
	timers.From(ctx).Tree(func(timer timers.Timer, depth int, _ *timers.TimerSet) {
		fmt.Printf("%s %s\n", strings.Repeat(" ", depth), timer.String())
	})

}

// Other examples

func ExampleTimerSet_New() {
	// The method chaining provides for a very useful way of performing timing on a
	// whole function. Assume your function is provided a context that has been generated
	// like so:
	ctx := context.Background()
	// At the top of the function we put this one line, and the "My Function" timer in
	// the current ctx now has the duration of this function.
	defer timers.From(ctx).New("My function").Start().Stop()

	// The New function supports string formatting directives using fmt.Sprintf
	defer timers.From(ctx).New("My function(%s, %d)", "Calling Parameter", 5)
}
func ExampleTimer_Tags() {
	// Grab a floating timer from the background context, which has no timerset
	timer := timers.From(context.Background()).New("MyFunc")
	// Assign it a tag and start
	timer.Tag("Test").Start()
	fmt.Printf("%s", timer.Tags())
	// Output:
	// [Test]
}

func goDoSomeWork(_ context.Context) {
	//Work Achieved
}

func ExampleNewContextWithTimer() {
	// Create a context to use for timings.
	ctx := context.Background()
	// Create a new TimerSet for grouping timers.
	newCtx, t := timers.NewContextWithTimer(ctx, "group")
	t.Start()
	goDoSomeWork(newCtx)
	t.Stop()
	workTimerSet := timers.From(newCtx)
	fmt.Printf("Work took %d ms and produced %d timers", t.Duration(), len(workTimerSet.All()))
}

func ExampleTimerSet_Wrap() {
	ctx := timers.NewContext(context.Background())
	timers.From(ctx).New("Stuff").Start().Stop()
	timers.From(ctx).Wrap(ctx, "Some Work", func(ctx context.Context) {
		timers.From(ctx).New("work step 1").Start().Stop()
		timers.From(ctx).New("work step 2").Start().Stop()
		timers.From(ctx).New("work step 3").Start().Stop()
	})

}

func ExampleTimerSet_MarshalJSON() {
	ctx := timers.NewContext(context.Background())
	timers.From(ctx).New("Timer 1")
	timers.From(ctx).New("Timer 1")
	timers.From(ctx).New("Timer 1")
	bytes, _ := json.MarshalIndent(timers.From(ctx), "", " ")
	fmt.Print(string(bytes))
	// Output:
	//[
	//  {
	//   "name": "Timer 1",
	//   "start": 0,
	//   "duration": 0
	//  },
	//  {
	//   "name": "Timer 1",
	//   "start": 0,
	//   "duration": 0
	//  },
	//  {
	//   "name": "Timer 1",
	//   "start": 0,
	//   "duration": 0
	//  }
	// ]
}
