// Copyright 2022 Nick Clifford <zaf@crypto.geek.nz>

/*
This package timers provides useful timers for measuring code paths in a production type
environment. While pprof may be considered an overkill or disadventagous for use
in production systems, this code is designed to be lightweight and produce useful
output.

It's original purpose was to aid in providing latency timings for downstream API
calls for an API service. With strategic timers placed around and use of middleware
performance timers for services can be provided. */
package timers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// A TimerSet is the list of timers inside a Context. The TimerSet provides a collection
// of functions to create, retrieve, and export timers. Creation of a TimerSet is done with
// the NewContext function.
type TimerSet struct {
	mu     sync.Mutex
	timers []*Timer
}

// An individual timer is used to measure, well, time elapsed, and is stored in a timerset.
// This type provides different functionality from a traditional time.Timer, as it has an
// interface designed to be quick and easy to instrument functions and code blocks.
//
// To get a Timer, you call New from a TimerSet, retrieved from the context:
//  timers.From(ctx).New("Blah")
type Timer struct {
	name     string
	start    time.Time
	duration time.Duration
	tags     []string
	subtimer *TimerSet
	// For export use only -- not at all guarenteed accurate except as copies being
	// generated for exporting tree
	id       int
	parentId int
}

type timerctx string

var GlobalTimers *TimerSet = &TimerSet{}

// Returns a context containing a new TimerSet. If the existing supplied context includes a
// TimerSet, then a timer is added to that with the name "Subtimer" to keep the timer tree.
// If you want to control the name of the parent timer, use NewContextWithTimer, if you
// do not want a timer at all, use context.Background(), or simply create a TimerSet empty
// struct that will not be attached to the current context.
func NewContext(ctx context.Context) context.Context {
	existingSet := From(ctx)
	newSet := &TimerSet{}
	ctx = context.WithValue(ctx, timerctx("timers"), newSet)
	t := existingSet.New("Subtimer")
	t.subtimer = newSet
	return ctx
}

// Returns a new context with a new TimerSet attached to it, and a timer attached to the previous
// context (if any). If the previous context did not have a TimerSet, then the timer is a floating
// timer (but safe to use). The timer name string can be a formatted string (just like fmt.Printf)
//
// TimerSets are threadsafe, you can create new timers across threads using the same TimerSet,
// however the Timers themselves are not. Don't manipulate the same Timer in different go routines
// at the same time. You're probably doing it wrong if you find yourself wanting to.
func NewContextWithTimer(ctx context.Context, name string, a ...interface{}) (context.Context, *Timer) {
	existingSet := From(ctx)
	newSet := &TimerSet{}
	ctx = context.WithValue(ctx, timerctx("timers"), newSet)
	t := existingSet.New(name, a...)
	t.subtimer = newSet
	return ctx, t
}

// Returns the TimerSet in the supplied context. If one doesn't exist, it returns a floating
// TimerSet that will be garbage collected. This ensures that chaining is always possible.
//
// Example:
//  t := timers.From(ctx).New("blah")
//
// Regardless of whether there is a TimerSet in the current context, you will get back a
// working timer you can use.
func From(ctx context.Context) *TimerSet {
	s := SetFromContext(ctx)
	if s == nil {
		s = newSet()
	}
	return s
}

// Deprecated: Use `From` instead, this function will disappear next major version.
func Get(ctx context.Context) *TimerSet {
	return From(ctx)
}

// Internal function to create a new timer from context.Background()
func newSet() *TimerSet {
	return &(TimerSet{})
}

// Get TimerSet from provided context (if any), otherwise return nil.
//
// You probably don't want this function. You want to always create a context with a new TimerSet
// (timers.NewContext()) and just use timers.From(ctx) to retrieve it -- knowing it's safe because
// if you end up in a context that has no timers it will still give you a working TimerSet.
// Otherwise what are you going to do if this function returns nil? Create a new context with
// NewContext()? Then just use NewContext in the first place.
func SetFromContext(ctx context.Context) *TimerSet {
	v := ctx.Value(timerctx("timers"))
	if v != nil {
		if t, ok := v.(*TimerSet); ok {
			return t
		}
	}
	return nil
}

// Returns a copy of all timers in this context. Note: These are a copy of the timers, not the
// original. TimerSets are threadsafe, Timers aren't.
func (s *TimerSet) All() []Timer {
	s.mu.Lock()
	defer s.mu.Unlock()
	timers := make([]Timer, len(s.timers))
	for i := 0; i < len(s.timers); i++ {
		timers[i] = *s.timers[i]
	}
	return timers
}

// Returns a copy of all timers in this context and timers created
// in child contexts (regardless of whether the underlying context)
// has been canceled.
func (s *TimerSet) AllDeep() []*Timer {
	timers, _ := s.flatTree(0, 1)
	return timers
}

// Flattens the tree, making a copy of it, setting the parentId and currentId for
// timers as it goes.
func (s *TimerSet) flatTree(parentId, currentId int) ([]*Timer, int) {
	srcTimers := s.All()
	timers := make([]*Timer, len(srcTimers))
	for i := 0; i < len(srcTimers); i++ {
		t := srcTimers[i]
		t.parentId = parentId
		t.id = currentId
		if t.subtimer != nil {
			var subtimers []*Timer
			subtimers, currentId = t.subtimer.flatTree(currentId, currentId+1)
			timers = append(timers, subtimers...)
		}
		currentId++
		timers[i] = &t
	}
	return timers, currentId
}

// Walk the timer tree. Since you can create new TimerSets in new contexts, the timers are
// effectively a tree. You can retrieve all the timers from the tree using GetAll() or you
// can walk the tree using Tree().
// This function takes a callback function that will be called with a COPY of each timer,
// and supplied with the current tree depth, and the parent TimerSet (that is the parent of
// the current TimerSet this Timer belongs to). If the timer is part of this TimerSet, then
// the parent will be nil.
func (s *TimerSet) Tree(fn func(Timer, int, *TimerSet)) {
	s.walkTree(nil, 0, fn)
}

// Internal function to walk the tree.
func (s *TimerSet) walkTree(p *TimerSet, depth int, fn func(Timer, int, *TimerSet)) {
	// We need a copy of all the timers, to avoid locking the entire tree!
	timers := s.All()
	// Now we are safe. We are doing a readonly.
	for _, t := range timers {
		fn(t, depth, p)
		if t.subtimer != nil {
			t.subtimer.walkTree(s, depth+1, fn)
		}
	}
}

// Wrap a block of work with a timer set in a new context. All timers that occur in the provided
// context will be children of the current context. The current context will have a timer
// measuring the duration.
//
// NOTE: You may find TImer.Wrap() to be better suited, it is simplier and doesn't require context
// passing.
//
// e.g. In the example below, a new Main context is created, and it will have 2 timers, Stuff and
// Some Work. The "Some Work" timer will have 3 sub timers under it.
//
func (s *TimerSet) Wrap(ctx context.Context, name string, fn func(context.Context)) {
	t := s.New(name)
	ns := &TimerSet{}
	newCtx := context.WithValue(ctx, timerctx("timers"), ns)
	t.subtimer = ns
	t.Start()
	fn(newCtx)
	t.Stop()
}

func (s *TimerSet) String() string {
	timers := s.AllDeep()
	var str []string
	for _, t := range timers {
		str = append(str, t.String())
	}
	return strings.Join(str, "\n")
}

// Create a new timer with the provided name.
// Name is a format string (like Printf)
func (s *TimerSet) New(name string, a ...interface{}) *Timer {
	name = fmt.Sprintf(name, a...)
	timer := Timer{
		name: name,
	}
	s.mu.Lock()
	s.timers = append(s.timers, &timer)
	s.mu.Unlock()
	return &timer
}

// Create a new global timer with the provided name.
// Name is a format string (like Printf)
// This is a convienance function for timers.GlobalTimers.New(...)
func New(name string, a ...interface{}) *Timer {
	name = fmt.Sprintf(name, a...)
	timer := Timer{
		name: name,
	}
	GlobalTimers.mu.Lock()
	GlobalTimers.timers = append(GlobalTimers.timers, &timer)
	GlobalTimers.mu.Unlock()
	return &timer
}

// Retrives the first timer with the provided name
// (Names do not have to be unique)
func (s *TimerSet) Find(name string) *Timer {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.timers {
		if t.name == name {
			return t
		}
	}
	return nil
}

// Starts the timer.
func (t *Timer) Start() *Timer {
	t.start = time.Now()
	return t
}

// Stops the timer
func (t *Timer) Stop() *Timer {
	t.duration = time.Since(t.start)
	return t
}

// internal, allows oneline sleep for testing
func (t *Timer) nap() *Timer {
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
	return t
}

// Returns how long the timer ran for.
// If the timer hasn't started, returns 0.
// If the timer is still running, it returns it's current runtime
// If the timer has been stopped it returns it's duration.
func (t *Timer) Duration() time.Duration {
	if t.duration != 0 {
		return t.duration
	} else if t.start.IsZero() {
		return 0
	} else {
		return time.Since(t.start)
	}
}

// Returns the duration the timer ran for or has been running in milliseconds, as a float rounded
// to 3 decimal places.
func (t *Timer) Milliseconds() float64 {
	// Microservers / 1000 is milliseconds, duration.Microseconds() returns an int, so when
	// the number is divided by 1000 (floating point rounding aside) it will round to 3 decimal
	// places.
	return float64(t.Duration().Microseconds()) / float64(1000)
}

// Returns true if the timer is  running
func (t *Timer) IsRunning() bool {
	return !t.start.IsZero() && t.duration == 0
}

// Tags the timer with a string. Multiple tags are supported.
// You can do timer.Timers(ctx).New("Test").Tag("tagA").Tag("tagB").Start()
func (t *Timer) Tag(tag string) *Timer {
	t.tags = append(t.tags, tag)
	return t
}

// Returns a list of all tags the timer has
func (t *Timer) Tags() []string {
	tags := make([]string, len(t.tags))
	copy(tags, t.tags)
	return tags
}

// Returns a _copy_ of the timers under this timer, if any. Returns empty list otherwise.
func (t *Timer) Children() []Timer {
	if t.subtimer == nil {
		return []Timer{}
	}
	return t.subtimer.All()
}

// Returns a string representing the timer's current state
func (t Timer) String() string {
	tags := ""
	if len(t.tags) > 0 {
		tags = fmt.Sprintf(" tags:(%s)", strings.Join(t.tags, ","))
	}
	if t.start.IsZero() {
		return fmt.Sprintf("%s: NotStarted%s", t.name, tags)
	} else if t.duration == 0 {
		return fmt.Sprintf("%s: Running%s", t.name, tags)
	} else {
		return fmt.Sprintf("%s: %.3fms%s", t.name, float64(t.duration)/float64(time.Millisecond), tags)
	}
}

// Compares the provided timer with this timer and returns true or false on whether they are equivilent.
// (Note: The timers have nanosecond resolution, so this function really is "is this the exact same
// timer?")
func (t *Timer) Compare(t2 *Timer) bool {
	if t.name == t2.name && t.start == t2.start && t.duration == t2.duration {
		return true
	} else {
		return false
	}
}

// Very trivial wrapper function. Simply wraps the provided function with the timer, allowing simple
// duration measurement. Use with closure style.
// Differs from timers.From(ctx).Wrap() because this method does not create a new
// context and TimerSet. See `TimerSet.Wrap()``
//  From(ctx).New("Timer").Wrap(func() {
//      SomeFunction(things, stuff)
//  })
func (t *Timer) Measure(fn func()) {
	t.Start()
	fn()
	t.Stop()
}

//  Marshaling type
type marshalTimer struct {
	Name     string          `json:"name"`
	Start    int64           `json:"start"`
	Duration float64         `json:"duration"`
	Tags     *[]string       `json:"tags,omitempty"`
	Children *[]marshalTimer `json:"children,omitempty"`
}

// Exports a TimerSet as a list of timers, each timer may have a
// "children" field, which contains a list of it's children in a
// tree like structure.
//
// To create an output suitable for a waterfall, use
// json.Marshal(set.GetAll())
func (s *TimerSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.toMarshalTimers())
}

func (s *TimerSet) toMarshalTimers() []marshalTimer {
	timers := make([]marshalTimer, len(s.timers))
	for i := 0; i < len(s.timers); i++ {
		timers[i] = s.timers[i].toMarshalTimer()
		if s.timers[i].subtimer != nil {
			list := s.timers[i].subtimer.toMarshalTimers()
			timers[i].Children = &list
		}
	}
	return timers
}

// Given a list of timers, turns it back into a TimerSet
// Be aware that the TimerSet will be lacking any context
// and will not be able to be associated to a context.
// You have essentially just imported a block of floating
// timers
func (s *TimerSet) UnmarshalJSON(bytes []byte) error {
	// We are given a list of Timers, hopefully.
	s.mu.Lock()
	err := json.Unmarshal(bytes, &s.timers)
	s.mu.Unlock()
	return err
}

func (t Timer) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.toMarshalTimer())
}
func (t *Timer) toMarshalTimer() marshalTimer {
	var tags *[]string
	if len(t.tags) > 0 {
		tags = &t.tags
	}
	var epoch = t.start.UnixMilli()
	if t.start.IsZero() {
		epoch = 0
	}
	return marshalTimer{
		Name:     t.name,
		Start:    epoch,
		Tags:     tags,
		Duration: t.Milliseconds(),
	}
}

func (t *Timer) UnmarshalJSON(bytes []byte) error {
	var mt marshalTimer
	err := json.Unmarshal(bytes, &mt)
	if err != nil {
		return err // Anyone know how I can get here in code coverage? :P
	}
	t.fromMarshaledTimer(mt)
	return nil
}

func (t *Timer) fromMarshaledTimer(mt marshalTimer) {
	t.name = mt.Name
	if mt.Start != 0 {
		t.start = time.UnixMilli(mt.Start)
	}
	if mt.Duration == 0 {
		// So here's a thing. If we're marshalling data from a file, if there is a zero
		// duration, it's probable that the actual millisecond value is zero, and not that
		// the timer hasn't stopped.
		// So we will interpret a Duration of zero, to be a duration of 1ns.
		t.duration = time.Duration(1)
	} else {
		t.duration = time.Duration(mt.Duration * float64(time.Millisecond))
	}
	if mt.Tags != nil {
		t.tags = *mt.Tags
	}
	if mt.Children != nil {
		s := newSet()
		t.subtimer = s
		s.timers = make([]*Timer, len(*mt.Children))
		for i := 0; i < len(*mt.Children); i++ {
			s.timers[i] = &Timer{}
			s.timers[i].fromMarshaledTimer((*mt.Children)[i])
		}
	}
}
