package timers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TimerId uint64

type TimerSet struct {
	mu     sync.Mutex
	timers []*Timer
	curId  TimerId
	ctx    context.Context
}

type Timer struct {
	name     string
	start    time.Time
	duration time.Duration
	tags     []string
	subtimer *TimerSet
}

type timerctx string

// Returns a new context with a TimerSet attached to it, and a timer that is attached
// to the previous context (if applicable -- if none, returns a floating timer which
// is safe to use)
func NewContext(ctx context.Context, name string, a ...interface{}) (context.Context, *Timer) {
	existingSet := Get(ctx)
	newSet := &TimerSet{}
	ctx = context.WithValue(ctx, timerctx("timers"), newSet)
	newSet.ctx = ctx
	t := existingSet.New(name, a...)
	t.subtimer = newSet
	return ctx, t
}

// Always returns a TimerSet, either from the provided
// context, or creates one on the fly that will be
// destroyed when it falls out of scope.
//
// This facilitates chaining. Eg
//    t := timers.Get(ctx).New("blah")
// Regardless of whether there is a timer set in the current
// context, you will get back a working timer you can use.
func Get(ctx context.Context) *TimerSet {
	v := ctx.Value(timerctx("timers"))
	if v != nil {
		if t, ok := v.(*TimerSet); ok {
			return t
		}
	}
	return newSet()
}
func newSet() *TimerSet {
	return &(TimerSet{ctx: context.Background()})
}

// Get TimerSet if in provided context, otherwise return nil
func GetFromContext(ctx context.Context) *TimerSet {
	v := ctx.Value(timerctx("timers"))
	if v != nil {
		if t, ok := v.(*TimerSet); ok {
			return t
		}
	}
	return nil
}

// Returns a copy of all timers in this context
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
func (s *TimerSet) AllDeep() []Timer {
	s.mu.Lock()
	defer s.mu.Unlock()
	var timers []Timer
	for _, t := range s.timers {
		timers = append(timers, *t)
		if t.subtimer != nil {
			timers = append(timers, t.subtimer.AllDeep()...)
		}
	}
	return timers
}

// Walk the timer tree
func (s *TimerSet) Tree(fn func(Timer, int)) {
	walkTree(s, 0, fn)
}

func walkTree(s *TimerSet, depth int, fn func(Timer, int)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.timers {
		fn(*t, depth)
		if t.subtimer != nil {
			walkTree(t.subtimer, depth+1, fn)
		}
	}
}

// Wrap a block of work with a timer set in a new context.
// All timers that occur in the provided context will be
// children of the current context. The current context will
// have a timer measuring the duration.
//
// eg In the example below, a new Main context is created, and
// it will have 2 timers, Stuff and Some Work. The "Some Work"
// timer will have 3 sub timers under it.
//
//    ctx, _ := timers.NewContext(ctx.Background(), "Main")
//    timers.Get(ctx).New("Stuff").Start().Stop()
//    timers.Get(ctx).Wrap("Some Work", func(ctx context.Context) {
//	      timers.Get(ctx).New("work step 1").Start().Stop()
//	      timers.Get(ctx).New("work step 2").Start().Stop()
//	      timers.Get(ctx).New("work step 3").Start().Stop()
//    })
func (s *TimerSet) Wrap(name string, fn func(context.Context)) {
	t := s.New(name)
	ns := &TimerSet{}
	ctx := context.WithValue(s.ctx, timerctx("timers"), ns)
	t.subtimer = ns
	t.Start()
	fn(ctx)
	t.Stop()
}

// Create a new timer with the provided name.
// Name is a format string (like Printf)
func (s *TimerSet) New(name string, a ...interface{}) *Timer {
	name = fmt.Sprintf(name, a...)
	timer := Timer{
		name: name,
	}
	s.mu.Lock()
	s.curId += 1
	s.timers = append(s.timers, &timer)
	s.mu.Unlock()
	return &timer
}

// Retrives the first timer with the provided name
// (Names do not have to be unique)
func (s *TimerSet) Get(name string) *Timer {
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
	return t.tags
}

// Returns a string representing the timer's current state
func (t Timer) String() string {
	if t.start.IsZero() {
		return fmt.Sprintf("%s: NotStarted", t.name)
	} else if t.duration == 0 {
		return fmt.Sprintf("%s: Running", t.name)
	} else {
		return fmt.Sprintf("%s: %.3fms", t.name, float64(t.duration)/float64(time.Millisecond))
	}
}
