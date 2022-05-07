# zafnz/go-timers
[![GoDoc](https://godoc.org/github.com/zafnz/go-timers?status.svg)](https://godoc.org/github.com/zafnz/go-timers)
[![CI](https://github.com/zafnz/go-timers/actions/workflows/main.yml/badge.svg)](https://github.com/zafnz/go-timers/actions/workflows/main.yml)

`go-timers` is a library to assist with timing various code blocks in go in such a way that the values can be 
inspected later, in runtime, without invoking a debugger.

It's original purpose was to track downstream API calls, to measure 
what downstream callers were impacting the handlers the most. 

# Usage

See `examples/main.go` for a full overview, however here are the highlights.

Timers are created and stored in a [context.Context](https://pkg.go.dev/context), this is because the typical
and expect use it they are associated with http transactions. Technically you can get the TimerSet and handle
it yourself, but that's not the idea. 

## Chaining and defer
Timers can be chained, which provides a very nice interaction.
```
t := timers.Get(ctx).New("My New Timer").Tag("mine").Start()
```
Due to the chaining they can be used to measure an entire function with just one line:
```
func MyFunction(ctx context.Context, args...) {
    defer timers.Get(ctx).New("MyFunction").Start().Stop()
    ...
}
```
This timer will terminate when the function returns. How neat!

## Grouping/children
Timers can be grouped by deriving a new context.
```
func MyFunc(ctx context.Context) {
    newCtx, _ := timers.NewContextWithTimer(ctx, "A new group of timers that are under the previous ctx")  
    timers.Get(newCtx).New("Timer1")
    timers.Get(newCtx).New("Timer2")
}
```

A wrap function provides a clean way to do this. 
```
func MyFunc(ctx context.Context) {
    ...
    // You have to provide the ctx to the Wrap function as the TimerSet itself
    // does not hold the context to avoid leaking it. 
    timers.Get(ctx).Wrap(ctx, "A library call", func(ctx context.Context) {
        MakeSomeCall(ctx, ...)
    })
    // All timers that MakeSomeCall() created are under "A library call" timer. 
    for _, t := range timers.Get(ctx).Get("A library call").Children {
      fmt.Println(t) // Prints all the timers that were created.
    }
}
```
 
## Thread safety

TimerSets are thread (go routine) safe, but Timers are not. Do not pass a Timer to another go routine, instead
pass the ctx and let it create it's own. 

