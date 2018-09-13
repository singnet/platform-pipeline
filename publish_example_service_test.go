package main

import (
	"errors"
	"github.com/DATA-DOG/godog"
	"log"
	"strconv"
	"strings"
)

var output_file string
var etherium_network_port int

func ethereum_network_is_running_on_port(port int) error {

	dir := env_singnet_repos + "/platform-contracts"
	output_file = log_path + "/ganache.log"
	etherium_network_port = port

	err := run_command(
		dir,
		output_file,
		"./node_modules/.bin/ganache-cli",
		"--mnemonic",
		"gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce")

	if err != nil {
		return err
	}

	exists, err := check_with_timeout(check_ethereium_network_is_running)
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
}

func check_ethereium_network_is_running() (bool, error) {

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

	success_string := "Listening on 127.0.0.1:" + strconv.Itoa(etherium_network_port)
	return strings.Contains(out, success_string), nil
}