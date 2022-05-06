package timers

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	timers := newSet()
	x := timers.New("Testing").Start()
	if timers.Get("Testing") == nil {
		t.Fatal("Failed to find created timer")
	}
	if x.IsRunning() == false {
		t.Error("Running timer isn't running")
	}
	time.Sleep(1 * time.Microsecond)
	timers.Get("Testing").Stop()
	if x.Duration() == 0 {
		t.Error("Failed to measure time passing")
	}

	if timers.Get("missing") != nil {
		t.Error("Get of a missing timer didn't return nil")
	}
}
func TestRunning(t *testing.T) {
	s := newSet()
	x := s.New("Running")
	if x.IsRunning() {
		t.Error("Unstarted timer says is running")
	}
	if x.Duration() != 0 {
		t.Error("Unstarted timer shows non zero duration")
	}
	if !strings.Contains(x.String(), "NotStarted") {
		t.Error("Unstarted timer's String() does not stay NotStarted")
	}
	x.Start()
	if !x.IsRunning() {
		t.Error("Running timer isn't running")
	}
	time.Sleep(1 * time.Microsecond)
	if x.Duration() == 0 {
		t.Error("Running timers duration is zero")
	}
	if !strings.Contains(x.String(), "Running") {
		t.Error("Running timer's String() does not say Running")
	}
	x.Stop()
	if x.IsRunning() {
		t.Error("Stopped timer says is still running")
	}
	t1 := x.Duration()
	time.Sleep(100 * time.Microsecond)
	t2 := x.Duration()
	if t1 != t2 {
		t.Error("Time has elapsed on stopped timer")
	}

	// Force duration to be 1 nanosecond
	x.duration = 5500 * time.Microsecond
	if !strings.Contains(x.String(), "5.5") {
		t.Error("timer's String() of 5.5ms doesn't contain 5.5")
	}
}

/*
func inSlice(s []string, n string) bool {
	for _, x := range s {
		if x == n {
			return true
		}
	}
	return false
}

func TestTags(t *testing.T) {
	s := NewSet()
	x := s.New("TagTest").Tag("A").Tag("B")
	tags := x.Tags()
	if len(tags) != 2 {
		t.Errorf("Have wrong number of tags (want 2, got %d)", len(tags))
	}
	if !inSlice(tags, "A") || !inSlice(tags, "B") {
		t.Error("Missing a tag")
	}
}*/

func TestNoContext(t *testing.T) {
	if Get(context.Background()) == nil {
		t.Error("Timers didn't return a TimerSet always")
	}
}
func TestContext(t *testing.T) {
	ctx := context.Background()
	ctx, _ = NewContext(ctx, "")
	if GetFromContext(ctx) == nil {
		t.Error("Failed to get timer in current context")
	}
	if GetFromContext(context.Background()) != nil {
		t.Error("Got a context from the background!?")
	}
	func(ctx context.Context) {
		defer Get(ctx).New("Test").Start().Stop()
		time.Sleep(100 * time.Microsecond)
		ctx, fn := context.WithCancel(ctx)
		Get(ctx).New("Deeper").Start().Stop()
		fn()
	}(ctx)
	if Get(ctx).Get("Test").Duration() < 100*time.Microsecond {
		t.Error("Time travel is impossible!")
	}
	if Get(ctx).Get("Deeper") == nil {
		t.Error("Timer lost in creation of new context from existing")
	}
	// Created a duplicate name
	Get(ctx).New("Test").Tag("blah")
	if len(Get(ctx).Get("Test").Tags()) != 0 {
		t.Error("Did not get first timer of duplicate name back")
	}
	timers := Get(ctx).All()
	if len(timers) != 3 {
		t.Error("Incorrect number of timers returned.")
	}

}

/*
func TestSubtimers(t *testing.T) {
	ctx := context.Background()
	ctx, _ = NewContext(ctx, "")
	x := Get(ctx).New("Step 1").Start()
	time.Sleep(1 * time.Millisecond)
	subTest(ctx)
	x.Stop()
	Get(ctx).New("Step 2").Start().Stop()
	subTest(ctx)
	Get(ctx).New("Step 3").Start()
	subTest(ctx)

	Get(ctx).Tree(func(timer Timer, depth int) {
		t.Logf("%s %s", strings.Repeat(" ", depth), timer.String())
	})
	t.Error("boop")
}

func TestWalk(t *testing.T) {

}
*/
