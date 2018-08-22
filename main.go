// The piper package enables the easy piping of output from one process into another.
// It tries to mimic the API of the 'os/exec' package from the stdlib.
package piper

import (
	"context"
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

// Chain holds a chain of commands where all output from a command is piped to the next one
type Chain struct {
	cmds []*exec.Cmd

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Allerr io.Writer
}

// Command creates a new Chain with the provided command as the first command.
// If behaves exactly like exec.Command but enables users to append more commands.
func Command(name string, arg ...string) *Chain {

	return &Chain{
		cmds: []*exec.Cmd{exec.Command(name, arg...)},
	}

}

// CommandContext creates a new Chain with the provided command as the first command
// If behaves exactly like exec.CommandContext but enables users to append more commands.
func CommandContext(ctx context.Context, name string, arg ...string) *Chain {

	return &Chain{
		cmds: []*exec.Cmd{exec.CommandContext(ctx, name, arg...)},
	}

}

// Cmd creates a new Chain with the provided command as the first command
// This function can be used when a more fine grained control over the process is
// necessary. You should not change the exec.Cmd after is has been added to the chain.
func Cmd(cmd *exec.Cmd) *Chain {

	return &Chain{
		cmds: []*exec.Cmd{cmd},
	}

}

// Command adds the command to the back of the command chain.
func (c *Chain) Command(name string, arg ...string) *Chain {

	c.cmds = append(c.cmds, exec.Command(name, arg...))
	return c

}

// CommandContext adds the command to the back of the command chain
func (c *Chain) CommandContext(ctx context.Context, name string, arg ...string) *Chain {

	c.cmds = append(c.cmds, exec.CommandContext(ctx, name, arg...))
	return c

}

func (c *Chain) Cmd(cmd *exec.Cmd) *Chain {

	c.cmds = append(c.cmds, cmd)
	return c

}

// CombinedOutput executes the chain and returns the combined output
func (c *Chain) CombinedOutput() ([]byte, error) {

	err := c.link()
	if err != nil {
		return nil, err
	}

	err = c.start()
	if err != nil {
		return nil, err
	}

	return c.cmds[len(c.cmds)-1].CombinedOutput()

}

func (c *Chain) Output() ([]byte, error) {

	err := c.link()
	if err != nil {
		return nil, err
	}

	err = c.start()
	if err != nil {
		return nil, err
	}

	return c.cmds[len(c.cmds)-1].Output()

}

func (c *Chain) Start() error {

	err := c.link()
	if err != nil {
		return err
	}

	err = c.start()
	if err != nil {
		return err
	}

	return c.cmds[len(c.cmds)-1].Start()

}

func (c *Chain) StdinPipe() (io.WriteCloser, error) {

	return c.cmds[0].StdinPipe()

}

func (c *Chain) StdoutPipe() (io.ReadCloser, error) {

	return c.cmds[len(c.cmds)-1].StdoutPipe()

}

func (c *Chain) StderrPipe() (io.ReadCloser, error) {

	return c.cmds[len(c.cmds)-1].StderrPipe()

}

func (c *Chain) Wait() error {

	var err error
	for i, cmd := range c.cmds {

		err = cmd.Wait()
		if err != nil {
			return errors.Wrapf(err, "unable to wait for process #%d (%s)", i, cmd.Path)
		}

	}

	return nil

}

func (c *Chain) link() error {

	for i := 0; i < len(c.cmds)-1; i++ {

		pipe, err := c.cmds[i].StdoutPipe()
		if err != nil {
			return errors.Wrapf(err, "unable to pipe command #%d (%s)", i, c.cmds[i].Path)
		}
		c.cmds[i+1].Stdin = pipe

		if c.Allerr != nil {
			c.cmds[i].Stderr = c.Allerr
		}

	}

	if c.Stdin != nil {
		c.cmds[0].Stdin = c.Stdin
	}
	if c.Stdout != nil {
		c.cmds[len(c.cmds)-1].Stdout = c.Stdout
	}
	if c.Stderr != nil {
		c.cmds[len(c.cmds)-1].Stderr = c.Stderr
	} else if c.Allerr != nil {
		c.cmds[len(c.cmds)-1].Stderr = c.Allerr
	}

	return nil

}

func (c *Chain) start() error {

	var err error
	for i := 0; i < len(c.cmds)-1; i++ {

		err = c.cmds[i].Start()
		if err != nil {
			errors.Wrapf(err, "unable to start command #%d (%s)", i, c.cmds[i].Path)
		}

	}

	return nil

}
