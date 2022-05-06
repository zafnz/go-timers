# Go-timers

A library to assist with timing various code blocks in go in such a way
that the values can be inspected later, in runtime, without invoking
a debugger.

It's original purpose was to track downstream API calls, to measure 
what downstream callers were impacting the handlers the most. 

# Usage

See `examples/main.go` for a good overview

