package logger

import (
	"bytes"
	"testing"

	"fmt"

	"encoding/json"

	"github.com/getsentry/raven-go"
	"github.com/mgutz/logxi/v1"
	"github.com/pkg/errors"
)

type mockTransport struct {
	Packet chan raven.Packet
}

func (t *mockTransport) Send(url, authHeader string, packet *raven.Packet) error {
	t.Packet <- *packet
	close(t.Packet)
	return nil
}

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
	if obj["_m"] != msg {
		t.Errorf("%v didn't match %v\n", obj["_m"], msg)
	}
	if obj["error"] != err.Error() {
		t.Errorf("%v didn't match %v\n", buf.String(), err.Error())
	}
	labels := obj["labels"].(map[string]interface{})
	if labels["one"] != "1" {
		t.Errorf("%v didn't match %v\n", labels["one"], "1")
	}
	if labels["two"] != "2" {
		t.Errorf("%v didn't match %v\n", labels["two"], 2)
	}
}

func TestSentry(t *testing.T) {
	client, err := raven.New("https://public:secret@sentry.example.com/1")
	if err != nil {
		t.Fatal(err)
	}
	mt := &mockTransport{
		Packet: make(chan raven.Packet),
	}
	client.Transport = mt
	raven.DefaultClient = client
	l := NewSentryLogger("test")
	l.Underlying().SetLevel(log.LevelFatal)

	l.Error(
		"Something is wrong",
		fmt.Errorf("This is an error"),
		"one", "1", "two", 2,
	)

	p := <-mt.Packet
	if len(p.Tags) == 0 {
		t.Error("There's supposed to be Tags")
	}
	if p.Logger != "test" {
		t.Error("The logger should be named test")
	}
	for _, tag := range p.Tags {
		switch k := tag.Key; k {
		case "one":
			if tag.Value != "1" {
				t.Errorf("Tag with key: %v should have value %v\n", k, "1")
			}
		case "two":
			if tag.Value != "2" {
				t.Errorf("Tag with key: %v should have value %v\n", k, "2")
			}
		default:
			t.Errorf("Unknown tag %v!\n", tag)
		}
	}

}
