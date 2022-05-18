package timers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Adds a Server-Timing header to the provided ResponseWriter, which contains all of the timers
// in the timerset, including all child timers.
//
// This function is useful primarily if you don't want to use the entire timers middleware suit
//
// Ex:
//  func ApiCall(w http.ResponseWriter, r *http.Request) {
//      ctx := timers.NewContext(r.Context())
//      result := DoAllTheApiWork(ctx) // Bunch of timers and subtimers created doing the work
//      timers.Get(ctx).AddHeader(w) // Add the response header
//      fmt.Fprintf(w, result)
//  }
func (s *TimerSet) AddHeader(w http.ResponseWriter) {
	timers := s.AllDeep()
	allValues := make([]string, len(timers))
	existing := make(map[string]struct{})
	for idx, timer := range timers {
		uniqueName := simplifyTimerName(existing, timer.name)
		allValues[idx] = timer.fmtAsHeader(uniqueName)
	}
	w.Header().Add("Server-Timing", strings.Join(allValues, ", "))
}

func (t *Timer) fmtAsHeader(uniqName string) string {
	if t.start.IsZero() {
		return fmt.Sprintf("%s;descr=%s;dur=0;parent=%d;id=%d", uniqName, quotedString(t.name), t.parentId, t.id)
	}
	return fmt.Sprintf("%s;descr=%s;dur=%.3f;start=%d;parent=%d;id=%d",
		uniqName, quotedString(t.name), t.Milliseconds(), t.start.UnixMilli(), t.parentId, t.id)
}

// Returns a timer name that doesn't exist.
func simplifyTimerName(existing map[string]struct{}, name string) string {
	pattern := regexp.MustCompile(`[^A-Za-z0-9_]`)
	name = pattern.ReplaceAllString(name, "_")
	if name == "" || name == "_" {
		name = "timer"
	}
	if _, ok := existing[name]; !ok {
		existing[name] = struct{}{}
		return name
	}
	for i := 0; ; i++ {
		tryName := name + strconv.Itoa(i)
		if _, ok := existing[tryName]; !ok {
			existing[tryName] = struct{}{}
			return tryName
		}
	}
}

// returns a quoted string where existing quotes have been escaped
func quotedString(str string) string {
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(str, "\"", "\\\""))
}
