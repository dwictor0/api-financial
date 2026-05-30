package tests

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("PIPEFY_TOKEN", "token_fake_teste_ci")
	os.Setenv("PIPEFY_PIPE_ID", "999999")
	os.Setenv("PIPEFY_API_URL", "https://api.pipefy.com")

	code := m.Run()

	os.Unsetenv("PIPEFY_TOKEN")
	os.Unsetenv("PIPEFY_PIPE_ID")
	os.Unsetenv("PIPEFY_API_URL")

	os.Exit(code)
}
