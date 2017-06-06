package utils

import (
	"context"
	"os/exec"
	"time"
)

func Command(timeout time.Duration, name string, arg ...string) (cmdSuccess bool, cmdOut []byte, err error) {
	if timeout <= 0 {
		timeout = 5
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, arg...)
	cmdOut, err = cmd.Output()
	if err != nil {
		return
	}
	cmdSuccess = cmd.ProcessState.Success()

	return
}
