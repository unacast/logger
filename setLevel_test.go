package logger

import (
	"testing"

	log "github.com/mgutz/logxi/v1"
)

func TestSetLevel(t *testing.T) {

	lgr := New("lgr")
	lgr2 := New("lgr2")

	SetLevel(log.LevelDebug)
	if !lgr.Underlying().IsDebug() || !lgr2.Underlying().IsDebug() {
		t.Error("Both the loggers should have Debug log level")
	}
	SetLevel(log.LevelInfo)
	if !lgr.Underlying().IsInfo() || !lgr2.Underlying().IsInfo() {
		t.Error("Both the loggers should have Info log level")
	}

}
