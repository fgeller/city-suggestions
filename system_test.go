package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
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

	url := url.URL{
		Scheme:   "http",
		Host:     addr,
		Path:     "suggestions",
		RawQuery: "q=SomeCityInTheMiddleOfNowhere",
	}

	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	awaitStartup(t, fmt.Sprintf("http://%v", addr), timeout)

	resp, err := http.Get(url.String())
	require.Nil(t, err, "should process request cleanly")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "should receive OK status code")
	require.Equal(t, contentTypeJSON, resp.Header.Get("Content-Type"), "should set content-type")

	bd, err := io.ReadAll(resp.Body)
	require.Nil(t, err, "should be able to read response body")
	require.JSONEq(t, `{"suggestions":[]}`, string(bd), "should return properly formatted JSON")
}

func TestQueryWok(t *testing.T) {
	build(t)

	addr := "0.0.0.0:7183"

	cmd := newCommand()
	err := cmd.start("./city-suggestions", "--addr", addr)
	require.Nil(t, err, "should start cleanly")

	url := url.URL{
		Scheme:   "http",
		Host:     addr,
		Path:     "suggestions",
		RawQuery: "q=Wok&latitude=43.70011&longitude=-79.4163",
	}

	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	awaitStartup(t, fmt.Sprintf("http://%v", addr), timeout)

	resp, err := http.Get(url.String())
	require.Nil(t, err, "should process request cleanly")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "should receive OK status code")
	require.Equal(t, contentTypeJSON, resp.Header.Get("Content-Type"), "should set content-type")

	bd, err := io.ReadAll(resp.Body)
	require.Nil(t, err, "should be able to read response body")
	expectedBody := `
{
  "suggestions": [
    {
      "name": "Wokingham",
      "latitude": "51.41120",
      "longitude": "-0.83565",
      "score": 0.9222222222222222
    },
    {
      "name": "Woking",
      "latitude": "51.31903",
      "longitude": "-0.55893",
      "score": 0.4416666666666667
    }
  ]
}
`
	require.JSONEq(t, expectedBody, string(bd), "should return properly formatted JSON")
}

func awaitStartup(t *testing.T, url string, ctx context.Context) {
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Error("startup timed out")
			return
		case <-tick.C:
			_, err := http.Get(url)
			if err == nil {
				return
			}
		}
	}
}

func build(t *testing.T) {
	var status int

	status, _, _ = newCommand().run("make", "build")
	require.Zero(t, status, "make build should succeed")

	status, _, _ = newCommand().run("ls", "city-suggestions")
	require.Zero(t, status, "ls notify should succeed")
}

type command struct {
	cmd *exec.Cmd
}

func newCommand() *command {
	return &command{}
}

func (c *command) start(name string, args ...string) error {
	c.cmd = exec.Command(name, args...)

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

	_ = c.cmd.Run()
	status := c.cmd.ProcessState.Sys().(syscall.WaitStatus)

	strOut := stdOut.String()
	strErr := stdErr.String()

	return status.ExitStatus(), strOut, strErr
}
