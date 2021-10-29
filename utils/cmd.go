package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/helloh2o/lucky/log"
	"os/exec"
)

// RunCmd 调用命令
func RunCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	log.Release("RUN CMD:: %s", cmd.String())
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Run()
	if err != nil {
		err = errors.New(fmt.Sprintf("RunCmd Error %s", string(b.Bytes())))
		return err
	}
	return err
}
