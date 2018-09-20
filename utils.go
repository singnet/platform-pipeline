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

type checkWithTimeoutType func() (bool, error)

// ExecCommand is used to run command from command line
type ExecCommand struct {
	Command    string
	Directory  string
	Env        []string
	Input      []string
	OutputFile string
	Args       []string
}

var envHome string
var envSingnetRepos string
var envGoPath string
var logPath string

func init() {

	envHome = os.Getenv("HOME")
	envSingnetRepos = os.Getenv("SINGNET_REPOS")
	envGoPath = os.Getenv("GOPATH")
	logPath = envGoPath + "/log"
	log.Printf("SINGNET_REPOS=%s\n", envSingnetRepos)
}

func readFile(file string) (string, error) {

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func fileExists(fileName string) (bool, error) {

	_, err := os.Stat(fileName)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func writeToFile(fileName string, content string) error {

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func appendToFile(fileName string, content string) error {

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func linkFile(fileName string, fileTo string) error {

	exists, err := fileExists(fileTo)

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	command := ExecCommand{
		Command: "ln",
		Args: []string{
			"-s",
			fileName,
			fileTo,
		},
	}

	return runCommand(command)
}

func runCommand(execCommad ExecCommand) error {

	log.Printf("[run_command] dir: '%s', command: '%s', args: '%s'\n",
		execCommad.Directory, execCommad.Command, strings.Join(execCommad.Args, ","))

	cmd, err := getCmd(execCommad)

	if err != nil {
		return err
	}

	return cmd.Run()
}

func runCommandAsync(execCommad ExecCommand) error {

	log.Printf("[run_command_async] dir: '%s', command: '%s', args: '%s'\n",
		execCommad.Directory, execCommad.Command, strings.Join(execCommad.Args, ","))

	cmd, err := getCmd(execCommad)

	if err != nil {
		return err
	}

	return cmd.Start()
}

func getCmd(execCommad ExecCommand) (*exec.Cmd, error) {

	cmd := exec.Command(execCommad.Command, execCommad.Args...)
	cmd.Dir = execCommad.Directory

	if execCommad.OutputFile != "" {
		stdOut, err := os.Create(execCommad.OutputFile)
		if err != nil {
			return nil, err
		}
		cmd.Stdout = stdOut
		cmd.Stderr = stdOut
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if len(execCommad.Input) > 0 {
		cmd.Stdin = strings.NewReader(strings.Join(execCommad.Input, "\n"))
	}

	cmd.Env = os.Environ()
	env := execCommad.Env
	if env != nil && len(env) > 0 {
		for _, e := range env {
			cmd.Env = append(cmd.Env, e)
		}
	}

	return cmd, nil
}

func checkWithTimeout(f checkWithTimeoutType) (bool, error) {
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

func toString(value int) string {
	return strconv.Itoa(value)
}
