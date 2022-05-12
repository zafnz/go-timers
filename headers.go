package timers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

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

func (s *TimerSet) AddHeader(w http.ResponseWriter) {
	timers := s.AllDeep()
	allValues := make([]string, len(timers))
	existing := make(map[string]struct{})
	for idx, timer := range timers {
		uniqueName := simplifyTimerName(existing, timer.name)
		allValues[idx] = timer.HeaderFmt(uniqueName)
	}
	w.Header().Add("Server-Timing", strings.Join(allValues, ", "))
}

func (t *Timer) HeaderFmt(uniqName string) string {
	if t.start.IsZero() {
		return fmt.Sprintf("%s;descr=%s;dur=0;parent=%d;id=%d", uniqName, quotedString(t.name), t.parentId, t.id)
	}
	return fmt.Sprintf("%s;descr=%s;dur=%.3f;start=%d;parent=%d;id=%d",
		uniqName, quotedString(t.name), t.Milliseconds(), t.start.UnixMilli(), t.parentId, t.id)
}
