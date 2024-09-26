package log

import (
	"errors"
	"testing"
)

func TestDebug(t *testing.T) {
	Debug("ok")
	Debug(errors.New("sm err"))
	Debug(struct {
		Name  string
		Value int
		Data  struct {
			HH string
		}
	}{
		"hell",
		534,
		struct{ HH string }{HH: "hh log"},
	})
	Debug("test format %v, %d, %s, %.2f", true, 1, "xx", 13.2222)
}
