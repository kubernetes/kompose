package utils

import (
	"bufio"
	"fmt"
	"os/exec"
)

// NewCommand returns an instance of exec.Cmd for any random command (containing pipes, etc)
func NewCommand(cmd string) *exec.Cmd {
	return exec.Command("sh", "-c", cmd)
}

// Execute wraps os/exec to execute shell commands and stream its output
func Execute(cmd *exec.Cmd) (output string, err error) {
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return output, err
	}

	scanner := bufio.NewScanner(cmdReader)
	text := ""
	go func() {
		for scanner.Scan() {
			text = scanner.Text()
			output = fmt.Sprintf("%s%s\n", output, text)
			fmt.Printf("%s\n", text)
		}
	}()

	err = cmd.Start()
	if err != nil {
		return output, err
	}

	err = cmd.Wait()
	if err != nil {
		return output, err
	}
	return output, err
}
