package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/zafnz/go-timers"
)

func otherMajorWork(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Println("Some other major work starting")
	time.Sleep(time.Duration(rand.Intn(1000) * int(time.Millisecond)))
	fmt.Println("Some other major work finished")
	wg.Done()

}
func moreMajorWork(ctx context.Context) {
	fmt.Println("more major work starting")
	time.Sleep(time.Duration(rand.Intn(1000) * int(time.Millisecond)))
	timers.Get(ctx).Wrap("subwork", func(ctx context.Context) {
		fmt.Println("Subwork going on")
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		fmt.Println("subwork finished")
	})
	fmt.Println("more major work finished")

}
func main() {
	ctx, _ := timers.NewContext(context.Background(), "main")
	// do things
	func() {
		defer timers.Get(ctx).New("Measure Sleep").Start().Stop()
		time.Sleep(100 * time.Millisecond)
	}()
	t := timers.Get(ctx).New("Count to a hundred million")
	t.Start()
	for i := 0; i < 100000000; i++ {
	}
	t.Stop()

	func() {
		ctx, t := timers.NewContext(ctx, "All major work")
		t.Start()
		defer t.Stop()

		wg := sync.WaitGroup{}
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go timers.Get(ctx).Wrap("majorWork", func(ctx context.Context) {
				otherMajorWork(ctx, &wg)
			})
		}
		wg.Wait()
		fmt.Println("All major work complete")
	}()

	workCtx, t := timers.NewContext(ctx, "More work")
	t.Start()
	moreMajorWork(workCtx)
	t.Stop()

	fmt.Println("\nLets see how we did:")
	for _, t := range timers.Get(ctx).AllDeep() {
		fmt.Println(t)
	}
	fmt.Println("\n\nAs a tree")
	timers.Get(ctx).Tree(func(timer timers.Timer, depth int) {
		fmt.Printf("%s %s\n", strings.Repeat(" ", depth), timer.String())
	})
}
