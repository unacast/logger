package logger

import (
	"context"
	"testing"
)

func TestRecoverPanics(t *testing.T) {
	ctx := context.Background()
	err := InitErrorReporting(ctx, "hepp", "test", "v1.0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		x := recover()
		if x != nil {
			t.Errorf("Didn't expect to recover here, but got %s", x)
		}
	}()
	defer ReportPanics(ctx)()
	defer CloseClient()
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
