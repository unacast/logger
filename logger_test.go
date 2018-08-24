package logger

import (
	"bytes"
	"testing"

	"encoding/json"

	"io/ioutil"
	"strings"

	"context"

	"github.com/mgutz/logxi/v1"
	"github.com/pkg/errors"
)

func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogger3(&buf, "test", log.NewJSONFormatter("test"))

	err := errors.New("This is an error")
	sl := unaLogger{
		Logger: l,
	}
	msg := "Something is wrong"
	sl.Error(
		msg,
		err,
		"one", "1", "two", 2,
	)

	var obj map[string]interface{}
	jsonErr := json.Unmarshal(buf.Bytes(), &obj)
	if jsonErr != nil {
		t.Fatalf("Hmm, couldn't unmarshal the log buffer %v. %v", buf.String(), jsonErr)
	}
	if obj["message"] != msg {
		t.Errorf("message \"%#v\" didn't match %#v\n", obj["message"], msg)
	}
	if obj["error"] != err.Error() {
		t.Errorf("error \"%v\" didn't match %v\n", buf.String(), err.Error())
	}
	if obj["one"] != "1" {
		t.Errorf("arg one \"%v\" didn't match %v\n", obj["one"], "1")
	}
	if obj["two"] != 2.0 {
		t.Errorf("arg two \"%v\" didn't match %v\n", obj["two"], 2.0)
	}
}

func TestErrorLoggingToFile(t *testing.T) {

	err := errors.New("This is an error")
	testLogFile := "/tmp/test.log"
	sl := NewLogger(Config{
		Name:     "test",
		FileName: testLogFile,
	})
	msg := "Something is wrong"
	sl.Error(
		msg,
		err,
		"one", "1", "two", 2,
	)

	fileContent, err := ioutil.ReadFile(testLogFile)
	if err != nil {
		t.Fatal("Didn't expect an error, got: ", err)
	}

	if !strings.Contains(string(fileContent), msg) {
		t.Errorf("Expected %v to contain %v, did not", testLogFile, msg)
	}

}

func TestLoggersInSeparateGoRoutines(t *testing.T) {
	go func() {
		lgr := New("lgr")
		lgr.Info("lgr")
	}()
	go func() {
		lgr2 := New("lgr2")
		lgr2.Info("lgr2")
	}()
}

func TestRecoverPanicsInFatal(t *testing.T) {
	ctx := context.Background()
	err := InitErrorReporting(ctx, "hepp", "test", "v1.0")
	if err != nil {
		t.Fatal(err)
	}
	defer ReportPanics(ctx)
	defer CloseClient()
	defer func() {
		x := recover()
		if x != nil {
			t.Errorf("there shouldn't be anything to recover here, but got %s", x)
		}
	}()

	lgr := New("unalogger")
	lgr.Fatal("FATAL", errors.New("FATAL ERROR"))
}
