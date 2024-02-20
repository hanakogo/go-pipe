package pipe

import (
	"bytes"
	"io"
	"os/exec"
)

// Command is an os/exec.Command wrapper for UNIX pipe
func Command(stdout *bytes.Buffer, stack ...*exec.Cmd) (err error) {
	var stderr bytes.Buffer
	defer stderr.Reset()
	pipeStack := make([]*io.PipeWriter, len(stack)-1)
	i := 0
	for ; i < len(stack)-1; i++ {
		inPipe, outPipe := io.Pipe()
		stack[i].Stdout = outPipe
		stack[i].Stderr = &stderr
		stack[i+1].Stdin = inPipe
		pipeStack[i] = outPipe
	}
	stack[i].Stdout = stdout
	stack[i].Stderr = &stderr

	return call(stack, pipeStack)
}

func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			closePipes(pipes)
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			closePipes(pipes)
			return err
		}
		defer func() {
			if err == nil {
				_ = pipes[0].Close()
				err = call(stack[1:], pipes[1:])
			} else {
				closePipes(pipes)
			}
		}()
	}
	err = stack[0].Wait()
	if err != nil {
		closePipes(pipes)
	}
	return err
}

func closePipes(pipes []*io.PipeWriter) {
	for _, pipe := range pipes {
		if pipe != nil {
			_ = pipe.Close()
		}
	}
}
