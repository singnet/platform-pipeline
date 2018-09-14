package main

import (
	"errors"
	"github.com/DATA-DOG/godog"
	"log"
	"strconv"
	"strings"
)

var output_file string
var output_contains_strings []string

var platform_contracts_dir string
var snet_config_file string

func init() {
	platform_contracts_dir = env_singnet_repos + "/platform-contracts"
	snet_config_file = env_home + "/.snet/config"
}

func ethereum_network_is_running_on_port(port int) error {

	output_file = log_path + "/ganache.log"
	output_contains_strings = []string{"Listening on 127.0.0.1:" + strconv.Itoa(port)}

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

func ipfs_is_runnig(port_api int, port_gateway int) error {

	env := []string{"IPFS_PATH=" + env_go_path + "/ipfs"}

	err := run_command("ipfs", "", "", env, "init")

	if err != nil {
		return err
	}

	err = run_command("ipfs", "", "", env, "bootstrap", "rm", "--all")

	if err != nil {
		return err
	}

	address_api := "/ip4/127.0.0.1/tcp/" + strconv.Itoa(port_api)
	err = run_command("ipfs", "", "", env, "config", "Addresses.API", address_api)

	if err != nil {
		return err
	}

	address_gateway := "/ip4/0.0.0.0/tcp/" + strconv.Itoa(port_gateway)
	err = run_command("ipfs", "", "", env, "config", "Addresses.Gateway", address_gateway)

	if err != nil {
		return err
	}

	output_file = log_path + "/ipfs.log"
	output_contains_strings = []string{
		"Daemon is ready",
		"server listening on " + address_api,
		"server listening on " + address_gateway,
	}
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

func identity_is_created_with_user_and_private_key(user string, private_key string) error {

	err := run_command("snet", "", "", nil,
		"identity", "create", user, "key", "--private-key", private_key)

	if err != nil {
		return err
	}

	err = run_command("snet", "", "", nil, "identity", "snet-user")

	return err
}

func snet_is_configured_with_ethereum_rpc_endpoint(endpoint_ethereum_rpc int) error {

	config := `
[network.local]
default_eth_rpc_endpoint = http://localhost:` + to_string(endpoint_ethereum_rpc)

	err := append_to_file(snet_config_file, config)

	if err != nil {
		return err
	}

	err = run_command("snet", "", "", nil, "network", "local")

	return err
}

func snet_is_configured_with_ipfs_endpoint(endpoint_ipfs int) error {

	config := `
[ipfs]
default_ipfs_endpoint = http://localhost:` + to_string(endpoint_ipfs)

	return append_to_file(snet_config_file, config)
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^Ethereum network is running on port (\d+)$`, ethereum_network_is_running_on_port)
	s.Step(`^Contracts are deployed using Truffle$`, contracts_are_deployed_using_truffle)
	s.Step(`^IPFS is running$`, ipfs_is_runnig)
	s.Step(`^IPFS is running with API port (\d+) and Gateway port (\d+)$`, ipfs_is_runnig)
	s.Step(`^Identity is created with user "([^"]*)" and private key "([^"]*)"$`,
		identity_is_created_with_user_and_private_key)
	s.Step(`^snet is configured with Ethereum RPC endpoint (\d+)$`, snet_is_configured_with_ethereum_rpc_endpoint)
	s.Step(`^snet is configured with IPFS endpoint (\d+)$`, snet_is_configured_with_ipfs_endpoint)

}

func check_daemon_is_running() (bool, error) {

	log.Printf("check output file: '%s'\n", output_file)
	log.Printf("check output file contains string: '%s'\n", strings.Join(output_contains_strings, ","))

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

	for _, str := range output_contains_strings {
		if !strings.Contains(out, str) {
			return false, nil
		}
	}

	return true, nil
}
