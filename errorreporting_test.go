package logger

import (
	"context"
	"testing"

	"strings"
)

func TestRecoverPanics(t *testing.T) {
	client, recoverPanics := SetUpErrorReporting(context.Background(), "hepp", "test", "v1.0")
	defer func() {
		x := recover()
		if !strings.Contains(x.(string), "Repanicked from logger") {
			t.Errorf("Expected 'Repanicked from logger' in the repanicked message. Was: %v", x)
		}
	}()
	defer recoverPanics()
	defer client.Close()

	panic("WOOT")
}
