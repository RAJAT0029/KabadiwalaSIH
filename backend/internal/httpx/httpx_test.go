package httpx

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRejectsExtraAndOversizedBodies(t *testing.T) {
	for _, body := range []string{`{"value":"x","owner":"forged"}`, `{"value":"x"}{}`, `{"value":"x"}` + strings.Repeat(" ", 1<<20)} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()
		var dst struct {
			Value string `json:"value"`
		}
		if Decode(w, r, &dst) {
			t.Fatal("invalid or oversized body accepted")
		}
	}
}
