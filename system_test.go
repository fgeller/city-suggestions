package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNoMatch(t *testing.T) {
	build(t)

	addr := "0.0.0.0:7183"

	cmd := newCommand()
	err := cmd.start("./city-suggestions", "--addr", addr)
	require.Nil(t, err, "should start cleanly")

	time.Sleep(20 * time.Millisecond)

	url := url.URL{
		Scheme:   "http",
		Host:     addr,
		Path:     "suggestions",
		RawQuery: "q=SomeCityInTheMiddleOfNowhere",
	}
	resp, err := http.Get(url.String())
	require.Nil(t, err, "should process request cleanly")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "should receive OK status code")
	require.Equal(t, contentTypeJSON, resp.Header.Get("Content-Type"), "should set content-type")

	bd, err := io.ReadAll(resp.Body)
	require.Nil(t, err, "should be able to read response body")
	require.JSONEq(t, `{"suggestions":[]}`, string(bd), "should return properly formatted JSON")
}

func build(t *testing.T) {
	var status int

	status, _, _ = newCommand().run("make", "build")
	require.Zero(t, status, "make build should succeed")

	status, _, _ = newCommand().run("ls", "city-suggestions")
	require.Zero(t, status, "ls notify should succeed")
}

type command struct {
	in  string
	cmd *exec.Cmd
	br  *blockingReader
}

func newCommand() *command {
	return &command{}
}

func (c *command) stdIn(in string) *command {
	c.in = in
	return c
}

type blockingReader struct {
	data    []byte
	release chan struct{}
}

func (r *blockingReader) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		var c int
		for ; c < len(p) && c < len(r.data); c++ {
			p[c] = r.data[c]
		}
		r.data = r.data[c:]
		return c, nil
	}

	<-r.release
	return 0, nil
}

func (c *command) blockStdIn(br *blockingReader) *command {
	c.br = br
	return c
}

func (c *command) start(name string, args ...string) error {
	c.cmd = exec.Command(name, args...)
	if len(c.in) > 0 {
		c.cmd.Stdin = strings.NewReader(c.in)
	} else if c.br != nil {
		c.cmd.Stdin = c.br
	}

	errPipe, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe err=%w", err)
	}

	go func() {
		for {
			ob := make([]byte, 4096)
			bc, err := errPipe.Read(ob)
			if err != nil && err != io.EOF {
				log.Printf("stderr pipe failed err=%v", err)
				return
			}
			if bc > 0 {
				log.Printf(">> stderr:\n%s\n", ob[:bc])
			}
		}
	}()

	outPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe err=%w", err)
	}
	go func() {
		for {
			ob := make([]byte, 4096)
			bc, err := outPipe.Read(ob)
			if err != nil && err != io.EOF {
				log.Printf("stdout pipe failed err=%v", err)
				return
			}
			if bc > 0 {
				log.Printf(">> stdout:\n%s\n", ob[:bc])
			}
		}
	}()

	return c.cmd.Start()
}

func (c *command) run(name string, args ...string) (int, string, string) {
	c.cmd = exec.Command(name, args...)

	var stdOut, stdErr bytes.Buffer
	c.cmd.Stdout = &stdOut
	c.cmd.Stderr = &stdErr

	if len(c.in) > 0 {
		c.cmd.Stdin = strings.NewReader(c.in)
	} else if c.br != nil {
		c.cmd.Stdin = c.br
	}

	_ = c.cmd.Run()
	status := c.cmd.ProcessState.Sys().(syscall.WaitStatus)

	strOut := stdOut.String()
	strErr := stdErr.String()

	return status.ExitStatus(), strOut, strErr
}

func (c *command) signal(sig os.Signal) error {
	return c.cmd.Process.Signal(sig)
}
