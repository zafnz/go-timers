// Unit tests for timers code. Extensive
package timers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	timers := newSet()
	x := timers.New("Testing").Start().nap()
	if timers.Find("Testing") == nil {
		t.Fatal("Failed to find created timer")
	}
	if x.IsRunning() == false {
		t.Error("Running timer isn't running")
	}
	timers.Find("Testing").Stop()
	if x.Duration() == 0 {
		t.Error("Failed to measure time passing")
	}

	if timers.Find("missing") != nil {
		t.Error("Find of a missing timer didn't return nil")
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

func TestWrap(t *testing.T) {
	ctx := NewContext(context.Background())
	From(ctx).Wrap(ctx, "Wrapped", func(c context.Context) {
		if c == ctx {
			t.Fatal("Identical context given to wrapped function")
		}
		if SetFromContext(c) == nil {
			t.Fatal("No TimerSet in new context")
		}
		time.Sleep(1 * time.Millisecond)
	})
	timer := From(ctx).Find("Wrapped")
	if timer == nil {
		t.Fatal("Wrap didn't create a wrapper timer in parent")
	}
	if timer.IsRunning() == true {
		t.Error("Wrap didn't stop timer")
	}
	if timer.Duration() == 0 {
		t.Error("Wrap didn't run timer")
	}
}

func TestNoContext(t *testing.T) {
	if From(context.Background()) == nil {
		t.Error("Timers didn't return a TimerSet always")
	}
}
func TestContext(t *testing.T) {
	ctx := context.Background()
	ctx = NewContext(ctx)
	if SetFromContext(ctx) == nil {
		t.Error("Failed to get timer in current context")
	}
	if SetFromContext(context.Background()) != nil {
		t.Error("Got a context from the background!?")
	}
}
func TestContextInheritence(t *testing.T) {
	ctx := NewContext(context.Background())

	From(ctx).New("Test").Start().nap().Stop()
	deeperCtx, fn := context.WithCancel(ctx)
	From(deeperCtx).New("Deeper").Start().nap().Stop()
	fn()

	if From(ctx).Find("Test").Duration() < 100*time.Microsecond {
		t.Fatal("Time travel is impossible -- I hope!")
	}
	if From(ctx).Find("Deeper") == nil {
		t.Error("Timer lost in creation of new context from existing")
	}
}

func TestTimerMeasure(t *testing.T) {
	ctx := NewContext(context.Background())
	From(ctx).New("TimerWrap").Measure(func() {
		// Simple stuff
		time.Sleep(10 * time.Nanosecond)
	})
	timer := From(ctx).Find("TimerWrap")
	if timer == nil {
		t.Fatal("Couldn't find TimerWrap")
	}
	if timer.Duration() == 0 {
		t.Error("Timer didn't measure anything")
	}
}
func TestDuplicateNames(t *testing.T) {
	set := newSet()
	t1 := set.New("timer")
	t2 := set.New("timer")
	t3 := set.Find("timer")
	if t3 != t1 {
		t.Error("Did not get the first duplicate")
	}
	for _, i := range set.timers {
		if i == t2 {
			return
		}
	}
	t.Error("Second duplicate not in timer list")
}

func TestCopy(t *testing.T) {
	set := newSet()
	t1 := set.New("t1")
	t2 := set.New("t2")
	t1.duration = 10
	t2.duration = 20

	all := set.All()
	// order should be preserved
	if all[0].duration != 10 && all[1].duration != 20 {
		t.Error("All() did not produce an identical copy")
	}
	if len(t1.Tags()) != 0 {
		t.Error("Where did that tag come from?")
	}
}

func TestCompare(t *testing.T) {
	t1 := newSet().New("blah")
	t2 := newSet().New("blah")
	t3 := newSet().New("blah").Start().Stop()
	if !t1.Compare(t2) {
		t.Error("Identical timers didn't compare correctly")
	}
	if t1.Compare(t3) {
		t.Error("Differening timers compared equal")
	}
}

const jsonBasicTimer = `{
	"name": "blah",
	"start": 1644884400000,
	"duration": 69000
}`

func TestTimerUnmarsha(t *testing.T) {
	var timer Timer
	err := json.Unmarshal([]byte(jsonBasicTimer), &timer)
	if err != nil {
		t.Fatal("Failed to unmarshal")
	}
	if timer.name != "blah" {
		t.Error("Failed to parse name")
	}
	if timer.duration.Seconds() != 69 {
		t.Error("Not nice")
	}
}
func TestTimerMarshal(t *testing.T) {
	t1 := newSet().New("Blah").Start().nap().Stop()
	bytes, err := json.Marshal(t1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
	if strings.Contains(string(bytes), "children") {
		t.Error("Timer with no children exported with children property")
	}
	// marshalling is lossy (saved in milliseconds)
	// so round t1 too.
	t1.start = time.UnixMilli(t1.start.UnixMilli())
	// Convert to milliseconds rounded to 3 decimal places (just like the marshall)
	ms := float64(t1.Duration().Microseconds()) / 1000
	t1.duration = time.Duration(ms * float64(time.Millisecond))
	var t2 Timer
	err = json.Unmarshal(bytes, &t2)
	if err != nil {
		t.Fatal(err)
	}
	if !t1.Compare(&t2) {
		t.Logf("%v", t2)
		bytes, err = json.Marshal(t1)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(bytes))
		t.Fatal("Marshalling and Unmarshalling didn't match")
	}

	t3 := newSet().New("Blah")
	bytes, err = json.Marshal(t3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
	err = json.Unmarshal(bytes, t3)
	if err != nil {
		t.Fatal(err)
	}
	if !t3.start.IsZero() {
		t.Error("Unstarted timer did not marshall correctly")
	}
}

func TestSetMarshalling(t *testing.T) {
	set1 := newSet()
	set1.New("t1").Start().nap().Stop()
	set1.New("t2").Start().nap().Stop()
	set1.New("t3").Start().nap().Stop()
	set1.New("t4").Start().nap().Stop()
	set1.New("t5").Start().nap().Stop()
	bytes, err := json.Marshal(set1)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(bytes), "\"t3\"") {
		t.Fatal("Marshall of set didn't contain all timers")
	}
	var set2 TimerSet
	err = json.Unmarshal(bytes, &set2)
	if err != nil {
		t.Fatal(err)
	}
	if len(set2.All()) != 5 {
		t.Error("Did not unmarshall all timers")
	}
}

func TestTreeMarshalling(t *testing.T) {
	depth0, _ := NewContextWithTimer(context.Background(), "depth0")
	From(depth0).New("t0.0").Start().nap().Stop()
	depth1, _ := NewContextWithTimer(depth0, "depth1")
	From(depth1).New("t1.0").Start().nap().Stop()
	depth2, _ := NewContextWithTimer(depth1, "depth2")
	From(depth2).New("t2.0").Start().nap().Stop()
	From(depth2).New("t2.1").Start().nap().Stop()
	depth3, _ := NewContextWithTimer(depth2, "depth3")
	From(depth3).New("3.0").Start().nap().Stop()
	From(depth3).New("3.1").Start().nap().Stop().Tag("tag3.1")
	t32 := From(depth3).New("3.2").Start().nap().Stop().Tag("3.2")

	t.Log("Tree")
	From(depth0).Tree(func(timer Timer, depth int, _ *TimerSet) {
		t.Logf("%s %s\n", strings.Repeat(" ", depth), timer.String())
	})

	bytes, err := json.MarshalIndent(From(depth0), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
	if !strings.Contains(string(bytes), "3.1") {
		t.Error("Marshalling all timers did not reveal a deep one")
	}

	var set TimerSet

	err = json.Unmarshal(bytes, &set)
	if err != nil {
		t.Fatal("Failed to unmarshal timer tree")
	}
	for _, timer := range set.AllDeep() {
		if timer.name == "3.2" && timer.duration.Milliseconds() == t32.duration.Milliseconds() {
			t.Log(timer)
			if len(timer.Tags()) != 1 {
				t.Fatalf("Timer didn't have tag %d", len(timer.Tags()))
			}
			if timer.Tags()[0] != "3.2" {
				t.Fatalf("Timer had wrong tag '%s'", timer.Tags()[0])
			}
			return
		}
	}
	t.Error("Did not discover timer 3.1")
}

func TestGlobalNew(t *testing.T) {
	timer := New("blah")
	timer.Start().Stop()
	timer = GlobalTimers.Find("blah")
	if timer == nil {
		t.Error("New() failed to create timer in GlobalTimers construct")
	}
}
