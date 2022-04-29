/*
 * The MIT License (MIT)
 *
 * Copyright © 2022 Hao Luo <haozzzzzzzz@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

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
