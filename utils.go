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

type checkWithTimeoutType func() (bool, error)

// ExecCommand is used to run command from command line
type ExecCommand struct {
	command         string
	directory       string
	env             []string
	input           []string
	outputFile      string
	args            []string
	async           bool
	err             error
	containsFile    string
	containsStrings []string
}

// NewCommand builder function
func NewCommand() *ExecCommand {
	return &ExecCommand{command: "sh"}
}

// Dir set dir
func (command *ExecCommand) Dir(dir string) *ExecCommand {
	command.directory = dir
	return command
}

// Output set output file
func (command *ExecCommand) Output(outputFile string) *ExecCommand {
	command.outputFile = outputFile
	return command
}

// CheckOutput set output file
func (command *ExecCommand) CheckOutput(strings ...string) *ExecCommand {
	return command.CheckFileContains(command.outputFile, strings...)
}

// CheckFileContains set output file
func (command *ExecCommand) CheckFileContains(file string, strings ...string) *ExecCommand {
	command.containsFile = file
	command.containsStrings = strings
	return command.checkContains()
}

// Run exec command
func (command *ExecCommand) Run(format string, a ...interface{}) *ExecCommand {
	command.async = false
	return command.run(format, a...)
}

// RunAsync exec command asynchroniously
func (command *ExecCommand) RunAsync(format string, a ...interface{}) *ExecCommand {
	command.async = true
	return command.run(format, a...)
}

func (command *ExecCommand) run(format string, a ...interface{}) *ExecCommand {

	if command.err != nil {
		return command
	}

	command.args = []string{"-c", fmt.Sprintf(format, a...)}

	if command.async {
		command.err = runCommandAsync(*command)
	} else {
		command.err = runCommand(*command)
	}

	return command
}

func (command *ExecCommand) checkContains() *ExecCommand {

	if command.err != nil || command.containsFile == "" {
		return command
	}

	fileContains := checkFileContains{
		output:  command.containsFile,
		strings: command.containsStrings,
	}

	if command.async {
		fmt.Println("check file async: ", command.containsFile, "strings: ", command.containsStrings)
		exists, err := checkWithTimeout(15000, 500, checkFileContainsStringsFunc(fileContains))
		command.err = err

		if err != nil && !exists {
			msg := fmt.Sprintf("file %s does not contains strings: %v",
				command.containsFile, command.containsStrings)
			command.err = errors.New(msg)
		}

	} else {
		ok, err := checkFileContainsStrings(fileContains)
		command.err = fileContainsError(fileContains, ok, err)
	}

	command.containsFile = ""

	return command
}

// Err returns error
func (command *ExecCommand) Err() error {
	err := command.err
	if err != nil {
		fmt.Println("command error: ", err.Error())
	}
	return err
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
		execCommad.directory, execCommad.command, strings.Join(execCommad.args, ","))

	cmd, err := getCmd(execCommad)

	fmt.Println("runCommand err: ", err)

	if err != nil {
		return err
	}

	return cmd.Run()
}

func runCommandAsync(execCommad ExecCommand) error {

	log.Printf("[run_command_async] dir: '%s', command: '%s', args: '%s'\n",
		execCommad.directory, execCommad.command, strings.Join(execCommad.args, ","))

	cmd, err := getCmd(execCommad)

	if err != nil {
		return err
	}

	return cmd.Start()
}

func getCmd(execCommad ExecCommand) (*exec.Cmd, error) {

	cmd := exec.Command(execCommad.command, execCommad.args...)
	cmd.Dir = execCommad.directory

	if execCommad.outputFile != "" {
		stdOut, err := os.Create(execCommad.outputFile)
		if err != nil {
			return nil, err
		}
		cmd.Stdout = stdOut
		cmd.Stderr = stdOut
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if len(execCommad.input) > 0 {
		cmd.Stdin = strings.NewReader(strings.Join(execCommad.input, "\n"))
	}

	cmd.Env = os.Environ()
	env := execCommad.env
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
		msg := fmt.Sprintf("File does not contain strings or has errors: %+v", fileContains)
		err = errors.New(msg)
		return
	}

	return
}
