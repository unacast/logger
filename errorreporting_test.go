package logger

import (
	"context"
	"testing"

	"strings"
)

func TestRecoverPanics(t *testing.T) {
	client, reportPanics := SetUpErrorReporting(context.Background(), "hepp", "test", "v1.0")
	defer func() {
		x := recover()
		if !strings.Contains(x.(string), "Repanicked from logger") {
			t.Errorf("Expected 'Repanicked from logger' in the repanicked message. Was: %v", x)
		}
	}()
	defer reportPanics()
	defer client.Close()

	panic("WOOT")
}

func TestInitErrorReporting(t *testing.T) {
	err := InitErrorReporting(context.Background(), "hepp", "test", "v1.0")
	if err != nil {
		t.Errorf("Didn't expect an error, but got %s", err)
	}
	if errorClient == nil {
		t.Errorf("Expected errorClient to be initialized, but it was nil")
	}
}
