package utils

import (
	"strings"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {
	name := "ping"
	args := []string{"-c", "2", "localhost"}

	cmdSuccess, out, err := Command(5*time.Second, name, args...)
	if err != nil {
		t.Fatal(err)
	}

	if !cmdSuccess {
		t.Fatal("'%v %v' command execute failed", name, strings.Join(args, " "))
	}

	//println(string(out))
	var _ = out
}

func TestCommandWithTimeout(t *testing.T) {
	name := "sleep"
	args := []string{"2"}

	_, _, err := Command(1*time.Second, name, args...)
	if err != nil {
		if !strings.Contains(err.Error(), "signal: killed") {
			t.Fatal(err)
		}
	}

}
