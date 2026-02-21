package cli

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/tasuku43/gion/internal/infra/debuglog"
)

type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) (stdout, stderr []byte, err error)
}

type realExecutor struct{}

func (realExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	trace := ""
	if debuglog.Enabled() {
		trace = debuglog.NewTrace("exec")
		debuglog.LogCommand(trace, debuglog.FormatCommand(name, args))
	}
	err := cmd.Run()
	if debuglog.Enabled() {
		debuglog.LogStdoutLines(trace, stdout.String())
		debuglog.LogStderrLines(trace, stderr.String())
		debuglog.LogExit(trace, debuglog.ExitCode(err))
	}
	return stdout.Bytes(), stderr.Bytes(), err
}

var defaultExecutor CommandExecutor = realExecutor{}
