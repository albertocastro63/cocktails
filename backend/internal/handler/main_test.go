package handler_test

import (
	"os"
	"testing"

	"github.com/almc/cocktails/internal/handler"
)

// TestMain disables the Forgot handler's timing-neutrality floor so the suite
// runs fast; the mitigation itself is exercised in production defaults.
func TestMain(m *testing.M) {
	handler.SetForgotFloor(0)
	os.Exit(m.Run())
}
