package main

import (
	"errors"
	"github.com/DATA-DOG/godog"
	"log"
	"strconv"
	"strings"
)

var output_file string
var output_contains_string string

var platform_contracts_dir string

func init() {
	platform_contracts_dir = env_singnet_repos + "/platform-contracts"
}

func ethereum_network_is_running_on_port(port int) error {

	output_file = log_path + "/ganache.log"
	output_contains_string = "Listening on 127.0.0.1:" + strconv.Itoa(port)

	err := run_command_async(
		"./node_modules/.bin/ganache-cli",
		platform_contracts_dir,
		output_file,
		nil,
		"--mnemonic",
		"gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce")

	if err != nil {
		return err
	}

	exists, err := check_with_timeout(check_daemon_is_running)
	if err != nil {
		return err
	}

	if !exists {
		errors.New("Etherium networks is not started!")
	}

	return nil
}

func contracts_are_deployed_using_truffle() error {

	err := run_command("./node_modules/.bin/truffle", platform_contracts_dir, "", nil, "compile")

	if err != nil {
		return err
	}

	err = run_command("./node_modules/.bin/truffle", platform_contracts_dir, "", nil, "migrate", "--network", "local")

	return err
}

func ipfs_is_runnig() error {

	env := []string{"IPFS_PATH=" + env_go_path + "/ipfs"}

	err := run_command("ipfs", platform_contracts_dir, "", env, "init")

	if err != nil {
		return err
	}

	err = run_command("ipfs", platform_contracts_dir, "", env, "bootstrap", "rm", "--all")

	if err != nil {
		return err
	}

	err = run_command("ipfs", platform_contracts_dir, "", env, "config", "Addresses.API", "/ip4/127.0.0.1/tcp/5002")

	if err != nil {
		return err
	}

	err = run_command("ipfs", platform_contracts_dir, "", env, "config", "Addresses.Gateway", "/ip4/0.0.0.0/tcp/8081")

	if err != nil {
		return err
	}

	output_file = log_path + "/ipfs.log"
	output_contains_string = "server listening on /ip4/0.0.0.0/tcp/8081"
	err = run_command_async("ipfs", platform_contracts_dir, output_file, env, "daemon")

	if err != nil {
		return err
	}

	exists, err := check_with_timeout(check_daemon_is_running)

	if err != nil {
		return err
	}

	if !exists {
		errors.New("Etherium networks is not started!")
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^Ethereum network is running on port (\d+)$`, ethereum_network_is_running_on_port)
	s.Step(`^Contracts are deployed using Truffle$`, contracts_are_deployed_using_truffle)
	s.Step(`^IPFS is running$`, ipfs_is_runnig)
}

func check_daemon_is_running() (bool, error) {

	log.Printf("check output file: '%s'\n", output_file)
	log.Printf("check output file contains string: '%s'\n", output_contains_string)

	out, err := read_file(output_file)
	if err != nil {
		return false, err
	}

	if out != "" {
		log.Printf("Output: %s\n", out)
	}

	if strings.Contains(out, "Error") {
		return false, errors.New("Output contains error!")
	}

	return strings.Contains(out, output_contains_string), nil
}