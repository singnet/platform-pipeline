package main

import (
	"errors"
	"fmt"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"log"
	"strconv"
	"strings"
)

var output_file string
var output_contains_strings []string

var platform_contracts_dir string
var example_service_dir string
var snet_config_file string

func init() {
	platform_contracts_dir = env_singnet_repos + "/platform-contracts"
	example_service_dir = env_singnet_repos + "/example-service"
	snet_config_file = env_home + "/.snet/config"
}

func ethereum_network_is_running_on_port(port int) error {

	output_file = log_path + "/ganache.log"
	output_contains_strings = []string{"Listening on 127.0.0.1:" + strconv.Itoa(port)}

	args := []string{"--mnemonic", "gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce"}
	command := ExecCommand{
		Command:    "./node_modules/.bin/ganache-cli",
		Directory:  platform_contracts_dir,
		OutputFile: output_file,
		Args:       args,
	}

	err := run_command_async(command)

	if err != nil {
		return err
	}

	exists, err := check_with_timeout(check_file_contains_strings)
	if err != nil {
		return err
	}

	if !exists {
		errors.New("Etherium networks is not started!")
	}

	return nil
}

func contracts_are_deployed_using_truffle() error {

	command := ExecCommand{
		Command:   "./node_modules/.bin/truffle",
		Directory: platform_contracts_dir,
		Args:      []string{"compile"},
	}

	err := run_command(command)

	if err != nil {
		return err
	}

	command.Args = []string{"migrate", "--network", "local"}
	err = run_command(command)

	return err
}

func ipfs_is_runnig(port_api int, port_gateway int) error {

	env := []string{"IPFS_PATH=" + env_go_path + "/ipfs"}

	command := ExecCommand{
		Command: "ipfs",
		Env:     env,
		Args:    []string{"init"},
	}

	err := run_command(command)

	if err != nil {
		return err
	}

	command.Args = []string{"bootstrap", "rm", "--all"}
	err = run_command(command)

	if err != nil {
		return err
	}

	address_api := "/ip4/127.0.0.1/tcp/" + strconv.Itoa(port_api)
	command.Args = []string{"config", "Addresses.API", address_api}
	err = run_command(command)

	if err != nil {
		return err
	}

	address_gateway := "/ip4/0.0.0.0/tcp/" + strconv.Itoa(port_gateway)
	command.Args = []string{"config", "Addresses.Gateway", address_gateway}
	err = run_command(command)

	if err != nil {
		return err
	}

	output_file = log_path + "/ipfs.log"
	command.OutputFile = output_file
	command.Args = []string{"daemon"}
	err = run_command_async(command)

	if err != nil {
		return err
	}

	output_contains_strings = []string{
		"Daemon is ready",
		"server listening on " + address_api,
		"server listening on " + address_gateway,
	}
	exists, err := check_with_timeout(check_file_contains_strings)

	if err != nil {
		return err
	}

	if !exists {
		errors.New("Etherium networks is not started!")
	}

	return nil
}

func identity_is_created_with_user_and_private_key(user string, private_key string) error {

	command := ExecCommand{
		Command: "snet",
		Args:    []string{"identity", "create", user, "key", "--private-key", private_key},
	}
	err := run_command(command)

	if err != nil {
		return err
	}

	command.Args = []string{"identity", "snet-user"}
	err = run_command(command)

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

	command := ExecCommand{
		Command: "snet",
		Args:    []string{"network", "local"},
	}
	err = run_command(command)

	if err != nil {
		return err
	}

	output_file = snet_config_file
	output_contains_strings = []string{"session"}
	exists, e := check_with_timeout(check_file_contains_strings)

	if !exists {
		return errors.New("snet config file is not created: " + snet_config_file)
	}

	return e
}

func snet_is_configured_with_ipfs_endpoint(endpoint_ipfs int) error {

	command := ExecCommand{
		Command: "sed",
		Args:    []string{"-ie", "/ipfs/,+2d", snet_config_file},
	}

	err := run_command(command)

	if err != nil {
		return err
	}

	config := `
[ipfs]
default_ipfs_endpoint = http://localhost:` + to_string(endpoint_ipfs)

	return append_to_file(snet_config_file, config)
}

func organization_is_added(table *gherkin.DataTable) error {

	organization := get_table_value(table, "organization")
	address := get_table_value(table, "address")
	member := get_table_value(table, "member")

	args := []string{
		"contract", "Registry",
		"--at", address,
		"createOrganization", organization,
		"[\"" + member + "\"]",
		"--transact",
	}

	command := ExecCommand{
		Command: "snet",
		Input:   []string{"y"},
		Args:    args,
	}

	return run_command(command)
}

func example_service_is_registered(table *gherkin.DataTable) error {

	name := get_table_value(table, "name")
	price := get_table_value(table, "price")
	endpoint := get_table_value(table, "endpoint")
	tags := get_table_value(table, "tags")
	description := get_table_value(table, "description")

	command := ExecCommand{
		Command:   "snet",
		Directory: example_service_dir,
		Input:     []string{"", "", name, "", price, endpoint, tags, description},
		Args:      []string{"service", "init"},
	}

	return run_command(command)
}

func example_service_is_published_to_network(table *gherkin.DataTable) error {

	agent_factory_address := get_table_value(table, "agent factory address")
	registry_address := get_table_value(table, "registry address")

	args := []string{
		"service", "publish", "local",
		"--config", "./service.json",
		"--agent-factory-at", agent_factory_address,
		"--registry-at", registry_address,
	}

	command := ExecCommand{
		Command:   "snet",
		Directory: example_service_dir,
		Input:     []string{"y", "y"},
		Args:      args,
	}

	return run_command(command)
}

func exampleservice_is_run_with_snet_daemon(table *gherkin.DataTable) error {

	daemon_port := get_table_value(table, "daemon port")
	ethereum_endpoint_port := get_table_value(table, "ethereum endpoint port")
	passthrough_endpoint_port := get_table_value(table, "passthrough endpoint port")

	agent_contract_address := get_table_value(table, "agent contract address")
	private_key := get_table_value(table, "private key")

	snetd_config_template := `
	{
    "AGENT_CONTRACT_ADDRESS": "%s",
    "AUTO_SSL_DOMAIN": "",
    "AUTO_SSL_CACHE_DIR": "",
    "BLOCKCHAIN_ENABLED": true,
    "CONFIG_PATH": "",
    "DAEMON_LISTENING_PORT": %s,
    "DAEMON_TYPE": "grpc",
    "DB_PATH": "./db",
    "ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:%s",
    "EXECUTABLE_PATH": "",
    "LOG_LEVEL": 5,
    "PASSTHROUGH_ENABLED": true,
    "PASSTHROUGH_ENDPOINT": "http://localhost:%s",
    "POLL_SLEEP": "",
    "PRIVATE_KEY": "%s",
    "SERVICE_TYPE": "jsonrpc",
    "SSL_CERT": "",
    "SSL_KEY": "",
    "WIRE_ENCODING": "json"
    }`

	snetd_config := fmt.Sprintf(snetd_config_template,
		agent_contract_address, daemon_port, ethereum_endpoint_port, passthrough_endpoint_port, private_key)

	file := example_service_dir + "/snetd.config.json"
	err := write_to_file(file, snetd_config)

	if err != nil {
		return err
	}

	link_file(env_singnet_repos+"/snet-daemon/build/snetd-linux-amd64", example_service_dir+"/snetd-linux-amd64")

	output_file = log_path + "/example-service.log"
	output_contains_strings = []string{}

	command := ExecCommand{
		Command:    example_service_dir + "/scripts/run-snet-service",
		Directory:  example_service_dir,
		OutputFile: output_file,
	}

	err = run_command_async(command)

	if err != nil {
		return err
	}

	_, err = check_with_timeout(check_file_contains_strings)

	if err != nil {
		return err
	}

	command = ExecCommand{
		Command:   example_service_dir + "/scripts/test-call",
		Directory: example_service_dir,
	}

	return run_command(command)
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
	s.Step(`^Organization is added:$`, organization_is_added)
	s.Step(`^example-service is registered$`, example_service_is_registered)
	s.Step(`^example-service is published to network$`, example_service_is_published_to_network)
	s.Step(`^example-service is run with snet-daemon$`, exampleservice_is_run_with_snet_daemon)

}

func check_file_contains_strings() (bool, error) {

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

func get_table_value(table *gherkin.DataTable, column string) string {

	names := table.Rows[0].Cells
	for i, cell := range names {
		if cell.Value == column {
			return table.Rows[1].Cells[i].Value
		}
	}

	log.Printf("column: %s has not been found in table", column)
	return ""
}
