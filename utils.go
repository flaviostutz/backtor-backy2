package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/sirupsen/logrus"
)

//ExecShellf execute shell command
func ExecShellf(command string, args ...interface{}) (string, error) {
	return ExecShellfTimeout(90*time.Second, command, args...)
}

//ExecShellTimeout execute shell command with timeout
func ExecShellfTimeout(timeout time.Duration, command string, args ...interface{}) (string, error) {
	command1 := fmt.Sprintf(command, args...)
	logrus.Debugf("shell command: '%s'", command1)
	acmd := cmd.NewCmd("bash", "-c", command1)
	statusChan := acmd.Start() // non-blocking
	running := true
	// if ctx != nil {
	// 	ctx.CmdRef = acmd
	// }

	//kill if taking too long
	if timeout > 0 {
		logrus.Debugf("Enforcing timeout %s", timeout)
		go func() {
			startTime := time.Now()
			for running {
				if time.Since(startTime) >= timeout {
					logrus.Warnf("Stopping command execution because it is taking too long (%d seconds)", time.Since(startTime))
					acmd.Stop()
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}

	// logrus.Debugf("Waiting for command to finish...")
	<-statusChan
	// logrus.Debugf("Command finished")
	running = false

	out := GetCmdOutput(acmd)
	status := acmd.Status()
	logrus.Debugf("shell output (%d): %s", status.Exit, out)
	if status.Exit != 0 {
		return out, fmt.Errorf("Failed to run command: '%s'; exit=%d; out=%s", command1, status.Exit, out)
	}
	return out, nil
}

//GetCmdOutput return content of executed command
func GetCmdOutput(cmd *cmd.Cmd) string {
	status := cmd.Status()
	out := strings.Join(status.Stdout, "\n")
	if len(status.Stderr) > 0 {
		if len(out) > 0 {
			out = out + "\n" + strings.Join(status.Stderr, "\n")
		} else {
			out = strings.Join(status.Stderr, "\n")
		}
	}
	return out
}
