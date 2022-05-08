package timers

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSimplifyTimerName(t *testing.T) {
	existing := make(map[string]struct{})
	name := "Test123_"
	v := simplifyTimerName(existing, name)
	if v != name {
		t.Errorf("'%s' should be allowed, got %s", name, v)
	}
	v = simplifyTimerName(existing, name)
	if v == name {
		t.Error("Duplicate simplifyTimerName allowed")
	}
}

func TestAddHeader(t *testing.T) {
	s := newSet()
	s.New("Test")
	timer := s.New("Test").Start().nap().Stop()
	response := httptest.NewRecorder()
	s.AddHeader(response)
	result := response.Result()
	header, found := result.Header["Server-Timing"]
	if !found || len(header) == 0 {
		t.Fatal("Header was not set")
	}
	txt := header[0]
	t.Log(txt)
	if !strings.Contains(txt, "\"Test\"") {
		t.Error("Test timer not present")
	}
	durStr := fmt.Sprintf("dur=%.3f", timer.Milliseconds())
	if !strings.Contains(txt, durStr) {
		t.Errorf("Header did not contain correct duration %s", durStr)
	}
}
