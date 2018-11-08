package main

import (
	"bufio"
	"errors"
	"fmt"
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

func fileExists(fileName string) bool {

	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
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

func linkFile(fileFrom string, fileTo string) error {

	if fileExists(fileTo) {
		return nil
	}

	return os.Symlink(fileFrom, fileTo)
}

func getPropertyFromFile(path string, property string) (string, error) {
	return getPropertyWithIndexFromFile(path, property, 0)
}

func getPropertyWithIndexFromFile(path string, property string, index int) (string, error) {

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)

	found := false
	for scanner.Scan() {

		if found {
			value := scanner.Text()
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return value, nil
		}

		if strings.Contains(scanner.Text(), property) {
			if index == 0 {
				found = true
			} else {
				index--
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errors.New("Property: " + property + " not found in file: " + path)
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

// timeout and tick are measurd in milliseconds units
func checkWithTimeout(timeout time.Duration,
	tick time.Duration,
	f checkWithTimeoutType) (bool, error) {

	_timeout := time.After(timeout * time.Millisecond)
	_tick := time.Tick(tick * time.Millisecond)

	for {
		select {
		case <-_timeout:
			return false, errors.New("timed out")
		case <-_tick:
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

type checkFileContains struct {
	strings    []string
	skipErrors []string
	ignoreCase bool
	output     string
}

func checkFileContainsStringsFunc(fileContains checkFileContains) checkWithTimeoutType {
	return func() (bool, error) {
		return checkFileContainsStrings(fileContains)
	}
}

func checkFileContainsStrings(fileContains checkFileContains) (bool, error) {

	if len(fileContains.skipErrors) > 0 {
		skipErrors := strings.Join(fileContains.skipErrors, ",")
		log.Printf("check output with skipped errors: '%s'\n", skipErrors)
	}

	out, err := readFile(fileContains.output)
	if err != nil {
		return false, err
	}

	if out != "" {
		log.Printf("Output: %s\n", out)
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(strings.ToLower(line), "error") {
			skip := false
			for _, skipErr := range fileContains.skipErrors {
				if strings.Contains(out, skipErr) {
					log.Printf("skipp error: '%s'\n", skipErr)
					skip = true
					break
				}
			}
			if !skip {
				return false, errors.New("Output contains error: '" + line + "'")
			}
		}
	}

	for _, str := range fileContains.strings {
		if !contains(out, str, fileContains.ignoreCase) {
			return false, nil
		}
	}

	return true, nil
}

func contains(str string, substr string, ignoreCase bool) bool {

	if ignoreCase {
		return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
	}

	return strings.Contains(str, substr)
}

func fileContainsError(fileContains checkFileContains, ok bool, e error) (err error) {

	if e != nil {
		err = e
		return
	}

	if !ok {
		msg := fmt.Sprintf("File contains error: %+v", fileContains)
		err = errors.New(msg)
		return
	}

	return
}
