package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

var outputFile string
var outputContainsStrings []string

var platformContractsDir string
var exampleServiceDir string
var snetConfigFile string

func init() {
	platformContractsDir = envSingnetRepos + "/platform-contracts"
	exampleServiceDir = envSingnetRepos + "/example-service"
	snetConfigFile = envHome + "/.snet/config"
}

func ethereumNetworkIsRunningOnPort(port int) error {

	outputFile = logPath + "/ganache.log"
	outputContainsStrings = []string{"Listening on 127.0.0.1:" + toString(port)}

	args := []string{"--mnemonic", "gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce"}
	command := ExecCommand{
		Command:    "./node_modules/.bin/ganache-cli",
		Directory:  platformContractsDir,
		OutputFile: outputFile,
		Args:       args,
	}

	err := runCommandAsync(command)

	if err != nil {
		return err
	}

	exists, err := checkWithTimeout(5000, 500, checkFileContainsStrings)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("Etherium networks is not started")
	}

	return nil
}

func contractsAreDeployedUsingTruffle() error {

	command := ExecCommand{
		Command:   "./node_modules/.bin/truffle",
		Directory: platformContractsDir,
		Args:      []string{"compile"},
	}

	err := runCommand(command)

	if err != nil {
		return err
	}

	command.Args = []string{"migrate", "--network", "local"}
	err = runCommand(command)

	return err
}

func ipfsIsRunning(portAPI int, portGateway int) error {

	env := []string{"IPFS_PATH=" + envGoPath + "/ipfs"}

	command := ExecCommand{
		Command: "ipfs",
		Env:     env,
		Args:    []string{"init"},
	}

	err := runCommand(command)

	if err != nil {
		return err
	}

	command.Args = []string{"bootstrap", "rm", "--all"}
	err = runCommand(command)

	if err != nil {
		return err
	}

	addressAPI := "/ip4/127.0.0.1/tcp/" + toString(portAPI)
	command.Args = []string{"config", "Addresses.API", addressAPI}
	err = runCommand(command)

	if err != nil {
		return err
	}

	addressGateway := "/ip4/0.0.0.0/tcp/" + toString(portGateway)
	command.Args = []string{"config", "Addresses.Gateway", addressGateway}
	err = runCommand(command)

	if err != nil {
		return err
	}

	outputFile = logPath + "/ipfs.log"
	command.OutputFile = outputFile
	command.Args = []string{"daemon"}
	err = runCommandAsync(command)

	if err != nil {
		return err
	}

	outputContainsStrings = []string{
		"Daemon is ready",
		"server listening on " + addressAPI,
		"server listening on " + addressGateway,
	}
	exists, err := checkWithTimeout(5000, 500, checkFileContainsStrings)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New("Etherium networks is not started")
	}

	return nil
}

func identityIsCreatedWithUserAndPrivateKey(user string, privateKey string) error {

	command := ExecCommand{
		Command: "snet",
		Args:    []string{"identity", "create", user, "key", "--private-key", privateKey},
	}
	err := runCommand(command)

	if err != nil {
		return err
	}

	command.Args = []string{"identity", "snet-user"}
	return runCommand(command)
}

func snetIsConfiguredWithEthereumRPCEndpoint(endpointEthereumRPC int) error {

	config := `
[network.local]
default_eth_rpc_endpoint = http://localhost:` + toString(endpointEthereumRPC)

	err := appendToFile(snetConfigFile, config)

	if err != nil {
		return err
	}

	command := ExecCommand{
		Command: "snet",
		Args:    []string{"network", "local"},
	}
	err = runCommand(command)

	if err != nil {
		return err
	}

	outputFile = snetConfigFile
	outputContainsStrings = []string{"session"}
	exists, e := checkWithTimeout(5000, 500, checkFileContainsStrings)

	if !exists {
		return errors.New("snet config file is not created: " + snetConfigFile)
	}

	return e
}

func snetIsConfiguredWithIPFSEndpoint(endpointIPFS int) error {

	command := ExecCommand{
		Command: "sed",
		Args:    []string{"-ie", "/ipfs/,+2d", snetConfigFile},
	}

	err := runCommand(command)

	if err != nil {
		return err
	}

	config := `
[ipfs]
default_ipfs_endpoint = http://localhost:` + toString(endpointIPFS)

	return appendToFile(snetConfigFile, config)
}

func organizationIsAdded(table *gherkin.DataTable) error {

	organization := getTableValue(table, "organization")
	address := getTableValue(table, "address")
	member := getTableValue(table, "member")

	args := []string{
		"contract", "Registry",
		"--at", address,
		"createOrganization", organization,
		"[\"" + member + "\"]",
		"--transact",
		"--yes",
	}

	command := ExecCommand{
		Command: "snet",
		Args:    args,
	}

	return runCommand(command)
}

func exampleserviceIsRegistered(table *gherkin.DataTable) error {

	name := getTableValue(table, "name")
	price := getTableValue(table, "price")
	endpoint := getTableValue(table, "endpoint")
	tags := getTableValue(table, "tags")
	description := getTableValue(table, "description")

	command := ExecCommand{
		Command:   "snet",
		Directory: exampleServiceDir,
		Input:     []string{"", "", name, "", price, endpoint, tags, description},
		Args:      []string{"service", "init"},
	}

	return runCommand(command)
}

func exampleserviceIsPublishedToNetwork(table *gherkin.DataTable) error {

	agentFactoryAddress := getTableValue(table, "agent factory address")
	registryAddress := getTableValue(table, "registry address")

	args := []string{
		"service", "publish", "local",
		"--config", "./service.json",
		"--agent-factory-at", agentFactoryAddress,
		"--registry-at", registryAddress,
		"--yes",
	}

	command := ExecCommand{
		Command:   "snet",
		Directory: exampleServiceDir,
		Args:      args,
	}

	return runCommand(command)
}

func exampleserviceIsRunWithSnetdaemon(table *gherkin.DataTable) error {

	daemonPort := getTableValue(table, "daemon port")
	ethereumEndpointPort := getTableValue(table, "ethereum endpoint port")
	passthroughEndpointPort := getTableValue(table, "passthrough endpoint port")

	agentContractAddress := getTableValue(table, "agent contract address")
	privateKey := getTableValue(table, "private key")

	snetdConfigTemplate := `
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

	snetdConfig := fmt.Sprintf(snetdConfigTemplate,
		agentContractAddress, daemonPort, ethereumEndpointPort, passthroughEndpointPort, privateKey)

	file := exampleServiceDir + "/snetd.config.json"
	err := writeToFile(file, snetdConfig)

	if err != nil {
		return err
	}

	linkFile(envSingnetRepos+"/snet-daemon/build/snetd-linux-amd64", exampleServiceDir+"/snetd-linux-amd64")

	outputFile = logPath + "/example-service.log"
	outputContainsStrings = []string{}

	command := ExecCommand{
		Command:    exampleServiceDir + "/scripts/run-snet-service",
		Directory:  exampleServiceDir,
		OutputFile: outputFile,
	}

	err = runCommandAsync(command)

	if err != nil {
		return err
	}

	_, err = checkWithTimeout(5000, 500, checkFileContainsStrings)

	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	command = ExecCommand{
		Command:   exampleServiceDir + "/scripts/test-call",
		Directory: exampleServiceDir,
	}

	return runCommand(command)
}

func singularityNETJobIsCreated(table *gherkin.DataTable) error {

	agentContractAddress := getTableValue(table, "agent contract address")
	maxPrice := getTableValue(table, "max price")

	args := []string{
		"agent",
		"--at", agentContractAddress,
		"create-jobs",
		"--funded",
		"--signed",
		"--max-price", maxPrice,
		"--yes",
	}

	command := ExecCommand{
		Command:   "snet",
		Directory: exampleServiceDir,
		Args:      args,
	}

	err := runCommand(command)

	if err != nil {
		return err
	}

	args = []string{
		"client", "call", "classify",
		fmt.Sprintf(`{"image_type": "jpg", "image": "%s"}`, testImage),
		"--agent-at", agentContractAddress,
	}

	command = ExecCommand{
		Command:   "snet",
		Directory: exampleServiceDir,
		Args:      args,
	}

	return runCommand(command)
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^Ethereum network is running on port (\d+)$`, ethereumNetworkIsRunningOnPort)
	s.Step(`^Contracts are deployed using Truffle$`, contractsAreDeployedUsingTruffle)
	s.Step(`^IPFS is running with API port (\d+) and Gateway port (\d+)$`, ipfsIsRunning)
	s.Step(`^Identity is created with user "([^"]*)" and private key "([^"]*)"$`,
		identityIsCreatedWithUserAndPrivateKey)
	s.Step(`^snet is configured with Ethereum RPC endpoint (\d+)$`, snetIsConfiguredWithEthereumRPCEndpoint)
	s.Step(`^snet is configured with IPFS endpoint (\d+)$`, snetIsConfiguredWithIPFSEndpoint)
	s.Step(`^Organization is added:$`, organizationIsAdded)
	s.Step(`^example-service is registered$`, exampleserviceIsRegistered)
	s.Step(`^example-service is published to network$`, exampleserviceIsPublishedToNetwork)
	s.Step(`^example-service is run with snet-daemon$`, exampleserviceIsRunWithSnetdaemon)
	s.Step(`^SingularityNET job is created$`, singularityNETJobIsCreated)

}

func checkFileContainsStrings() (bool, error) {

	log.Printf("check output file: '%s'\n", outputFile)
	log.Printf("check output file contains string: '%s'\n", strings.Join(outputContainsStrings, ","))

	out, err := readFile(outputFile)
	if err != nil {
		return false, err
	}

	if out != "" {
		log.Printf("Output: %s\n", out)
	}

	if strings.Contains(out, "Error") {
		return false, errors.New("Output contains error")
	}

	for _, str := range outputContainsStrings {
		if !strings.Contains(out, str) {
			return false, nil
		}
	}

	return true, nil
}

func getTableValue(table *gherkin.DataTable, column string) string {

	names := table.Rows[0].Cells
	for i, cell := range names {
		if cell.Value == column {
			return table.Rows[1].Cells[i].Value
		}
	}

	log.Printf("column: %s has not been found in table", column)
	return ""
}

var testImage = "/9j/4AAQSkZJRgABAQAAAQABAAD/2wCEAAkGBxAPDw0PDQ0PDg0PDQ0PDQ0PDQ8ODQ0NFRIWFhUSExUYHyghGB4lJxMWITEhJSor" +
	"Li4uGB8zODMsNygtLisBCgoKDQ0NDxAPFSsdFRktLSs4LCsrKysrKysrKzcrLSsrKy0tKzcrKysrKysrKystNy0rKysrKysrKysr" +
	"KysrK//AABEIAOEA4QMBIgACEQEDEQH/xAAcAAEAAQUBAQAAAAAAAAAAAAAAAQIDBAUGBwj/xABAEAEAAgECAgcEBAoLAQAAAAAA" +
	"AQIDBBEFIQYHEhMxQWFRcZGxMkKBghQiI3KDkqGissEIF1NUYpOjwtHh8BX/xAAWAQEBAQAAAAAAAAAAAAAAAAAAAQL/xAAWEQEB" +
	"AQAAAAAAAAAAAAAAAAAAEQH/2gAMAwEAAhEDEQA/APcQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
	"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
	"AEJavjHGKabsRe1K3vv2O3aKVnbbfn9scgbRG7jtV0g1XjSKxWfCa0i1fjzYF+OaqfHPMfmxWPkD0Dc3ee//AEs0/SzXn70q6ay0" +
	"+N7frSlHf7m7iseaZ+tPxlk4r+s/GVHWG7n8NLz4Tf8AWmGTSuSNvyto+9MiVuBqq58sfW329tYU243XHMRmmkbzFYmLecztHKRW" +
	"3EQkAAAAAAAAAAAAAABby56UibXvWlY8bWmKxH2y13SPjWPRae+fJtO3KlZtFe3f2b+UcpmZ8oiXzv0g41rOOavutPGfUT2prWkX" +
	"tGGN5+rjjlWvrbefOdvAH0Nk6T6Cv0tdp4/TUl551q8c0mpppq6fU481/wAtE1pbteVZ5uMp1McVmsTP4LW08+xOptvHpvFWB0g6" +
	"DcQ4Vp41GpjT1xfhGPfusk5LTNpiu2+0bRyhUYmk4jk0/b7jJOObREb18vWPZ4s7F0y1lYiLXrl287d5Fp98xZps/jLGmVSOjv03" +
	"1EzMzHYjs8opett7b+M9uszt6Ndbp7xKsztfBNd+W+Cm+3q02SWLkNV0tesjicf3f/Jj/lcjrO4p7dNH6H/tyMoQdl/Wjxfy1GGv" +
	"l+LhhTbrH4raPxtZbf8Aw1rVx6uAdFm6Wa/J9PWZ538fykx8mz6LZb5dXpu8vbJPfYtptabfXj2uQxO06vcXb12jr7dRh393bjcH" +
	"0qECKAAAAAAAAAAAAAA8O/pBcbvXLp9JWZikaeMt9vC1r3tG3+nHxbrqI4TSmnz55rE5N6Ui3nG9Ztaft3j4Of8A6RXDLxl0urrE" +
	"93fDGG0+UXpe1o+MX/Y3HUXxuk0y6e1oi2SKZMUT52rE1vX38o+Erg9ecV1xaLvuC62IjnSMeSPSa3iXaxLXdI9H3+j1WKY37eDJ" +
	"G3rtvHyQfLVrdqtbR9albfGIY9pXcNdsVYnxpNqT762mv8li0tIoySxrr1pWLoKJQlAEKoUK4BfxPSeqLS95r8M+VO1f3bVnb+Tz" +
	"XD4va+ozQ721OaY5Ux0pE+tpmZ/hQevwAKAAAAAAAAAAAAAA03Szo7i4lpMulz8otzx5Ij8bFlj6N4/94TL5u4jwniHANVtel+xF" +
	"+1iy07Vcd4ieVsd48J9PH5vqpj63RYs9Jx58VMuO0bWpkrF6z9kg8e4P14460rXV6bJa8RtN68pn37bxLbf138PmOeDUe6YiGx4t" +
	"1P8AC802tipl0lp/sctrUj3UvvEfY4/inUXkjedJrcWSee1NRinHM/epv8io4HUWre2ovjiYx31WovjiY2mMd8k2pvHumGtyOy4l" +
	"0O1ugwXnXY6xvetaXpkjJS0RTaNp8fKPGHH6iNpUY9lmy5ZasCiUJlAJhMKd1UAyNNG9ofS/VTwzuOG4rTG1s8zln29nwr8v2vn7" +
	"olwy2q1enwUjnkyRX3R5z9kRMvqzSYK4sePHSNqY6UpWPZWsbQir4jc3BIjdIAAAAAAAAAAAAAAAANV0m4VGs0ubBPjakzjn2ZI5" +
	"1l8x8b0VsOTJS9drVtatonymJ2fWEw8f65OjG141uGm9cnLP2a8q5IjlafZv84XDXit1qWTnpsxpEUShMqQSrqt7tl0f4Xk1eoxY" +
	"MUb3yWiPbFY352n0jxQeudRnAezGXX5K7bTOHTzPtmI7do+O3xevd40PCNPTS4MOnxRtjxUiseU2nztPrM82dXNuK2PeJ7xg1uuR" +
	"YGZFlcSxK2X6SC8IhIAAAAAAAAAAAAAACm9ItExaImJ8YmN4mFQDlOK9XfCtVM2y6GlbT9fDfJgn9yYhodR1M8Jn6P4VT0jUzaP3" +
	"ol6SpsDyfN1K6D6up1cfexz/ALWPHUvoonnqdVMe/HH8nrlqLU4geZYOqDhlfpVz5PztRav8OzpODdEtJo940unri35WtG83mPW0" +
	"zM/tdT3SYxA11NIvV07N7tVFAY1MK53a92U9kFqtF2tVUQkBKEgAAAAAAAAAAAAAAAAI2SApmDZUAp2TskBGxskBAkBAkAAAAAAA" +
	"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB//2Q=="
