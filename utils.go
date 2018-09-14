package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type check_with_timeout_type func() (bool, error)

type ExecCommand struct {
	Command    string
	Directory  string
	Env        []string
	Input      []string
	OutputFile string
	Args       []string
}

var env_home string
var env_singnet_repos string
var env_go_path string
var log_path string

func init() {

	env_home = os.Getenv("HOME")
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

func append_to_file(file_name string, content string) error {

	file, err := os.OpenFile(file_name, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func run_command(exec_commad ExecCommand) error {

	log.Printf("[run_command] dir: '%s', command: '%s', args: '%s'\n",
		exec_commad.Directory, exec_commad.Command, strings.Join(exec_commad.Args, ","))

	cmd, err := get_cmd(exec_commad)

	if err != nil {
		return err
	}

	return cmd.Run()
}

func run_command_async(exec_commad ExecCommand) error {

	log.Printf("[run_command_async] dir: '%s', command: '%s', args: '%s'\n",
		exec_commad.Directory, exec_commad.Command, strings.Join(exec_commad.Args, ","))

	cmd, err := get_cmd(exec_commad)

	if err != nil {
		return err
	}

	return cmd.Start()
}

func get_cmd(exec_commad ExecCommand) (*exec.Cmd, error) {

	cmd := exec.Command(exec_commad.Command, exec_commad.Args...)
	cmd.Dir = exec_commad.Directory

	if exec_commad.OutputFile != "" {
		std_out, err := os.Create(exec_commad.OutputFile)
		if err != nil {
			return nil, err
		}
		cmd.Stdout = std_out
		cmd.Stderr = std_out
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if len(exec_commad.Input) > 0 {
		cmd.Stdin = strings.NewReader(strings.Join(exec_commad.Input, "\n"))
	}

	cmd.Env = os.Environ()
	env := exec_commad.Env
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

func to_string(value int) string {
	return strconv.Itoa(value)
}
