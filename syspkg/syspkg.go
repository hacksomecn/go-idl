package syspkg

import (
	"context"
	"github.com/sirupsen/logrus"
	"os/exec"
	"syscall"
)

func RunCommand(
	workingDir string,
	name string,
	args ...string,
) (exit int, output string, err error) {
	execCommand := exec.Command(name, args...)
	if workingDir != "" {
		execCommand.Dir = workingDir
	}
	bOutput, err := execCommand.Output()

	if nil != err {
		logrus.Errorf("run `%s %s`failed.", name, args)
		exitError, ok := err.(*exec.ExitError)
		if ok {
			err = exitError
			logrus.Error(string(exitError.Stderr))
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			exit = waitStatus.ExitStatus()
		}
		return

	}

	output = string(bOutput)
	waitStatus := execCommand.ProcessState.Sys().(syscall.WaitStatus)
	exit = waitStatus.ExitStatus()
	return
}

func RunCommandCtx(
	ctx context.Context,
	workingDir string,
	name string,
	args ...string,
) (exit int, output string, err error) {
	execCommand := exec.CommandContext(ctx, name, args...)
	execCommand.Dir = workingDir
	bOutput, err := execCommand.Output()
	output = string(bOutput)

	if nil != err {
		logrus.Errorf("run `%s %s`failed.", name, args)
		exitError, ok := err.(*exec.ExitError)
		if ok {
			err = exitError
			logrus.Error(string(exitError.Stderr))
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			exit = waitStatus.ExitStatus()
		}
		return

	}

	output = string(bOutput)
	waitStatus := execCommand.ProcessState.Sys().(syscall.WaitStatus)
	exit = waitStatus.ExitStatus()

	return
}
