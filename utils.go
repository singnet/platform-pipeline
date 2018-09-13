package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type check_with_timeout_type func() (bool, error)

var env_singnet_repos string
var env_go_path string
var log_path string

func init() {

	env_singnet_repos = os.Getenv("SINGNET_REPOS")
	env_go_path = os.Getenv("GOPATH")
	log_path = env_go_path + "/log"
	log.Printf("SINGNET_REPOS=%s%\n", env_singnet_repos)
}

func read_file(file string) (string, error) {

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func run_command(command string, dir string, out string, env []string, args ...string) error {

	log.Printf("[run_command] dir: '%s', command: '%s', args: '%s'\n", dir, command, strings.Join(args, ","))

	cmd, err := get_cmd(command, dir, out, env, args...)

	if err != nil {
		return err
	}

	return cmd.Run()
}

func run_command_async(command string, dir string, out string, env []string, args ...string) error {

	log.Printf("[run_command_async] dir: '%s', command: '%s', args: '%s'\n", dir, command, strings.Join(args, ","))
	cmd, err := get_cmd(command, dir, out, env, args...)

	if err != nil {
		return err
	}

	return cmd.Start()
}

func get_cmd(command string, dir string, out string, env []string, args ...string) (*exec.Cmd, error) {

	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	if out != "" {
		std_out, err := os.Create(out)
		if err != nil {
			return nil, err
		}
		cmd.Stdout = std_out
		cmd.Stderr = std_out
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.Env = os.Environ()
	if env != nil && len(env) > 0 {
		for _, e := range env {
			cmd.Env = append(cmd.Env, e)
		}
	}

	return cmd, nil
}

func check_with_timeout(f check_with_timeout_type) (bool, error) {
	timeout := time.After(5 * time.Second)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return false, errors.New("timed out")
		case <-tick:
			ok, err := f()
			if err != nil {
				return false, err
			} else if ok {
				return true, nil
			}
		}
	}
}
