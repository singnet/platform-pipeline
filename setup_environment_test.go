package main

import (
	"errors"
	"log"

	"github.com/DATA-DOG/godog/gherkin"
)

var platformContractsDir string
var exampleServiceDir string
var dnnModelServicesDir string
var snetConfigFile string

var accountPrivateKey string
var identiyPrivateKey string

var agentFactoryAddress string
var singnetTokenAddress string
var registryAddress string
var multiPartyEscrow string
var organizationAddress string
var agentAddress string

var environmentIsSet = false
var serviceIsPublished = false

func init() {
	platformContractsDir = envSingnetRepos + "/platform-contracts"
	exampleServiceDir = envSingnetRepos + "/example-service"
	dnnModelServicesDir = envSingnetRepos + "/dnn-model-services/Services/gRPC/Basic_Template"
	snetConfigFile = envHome + "/.snet/config"
}

func ethereumNetworkIsRunningOnPort(port int) (err error) {

	if environmentIsSet {
		return
	}

	outputFile = logPath + "/ganache.log"
	outputContainsStrings = []string{"Listening on 127.0.0.1:" + toString(port)}

	args := []string{"--mnemonic", "gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce"}
	command := ExecCommand{
		Command:    "./node_modules/.bin/ganache-cli",
		Directory:  platformContractsDir,
		OutputFile: outputFile,
		Args:       args,
	}

	err = runCommandAsync(command)

	if err != nil {
		return
	}

	exists, err := checkWithTimeout(5000, 500, checkFileContainsStrings)
	if err != nil {
		return
	}

	if !exists {
		return errors.New("Etherium networks is not started")
	}

	organizationAddress, err = getPropertyFromFile(outputFile, "(1)")
	if err != nil {
		return
	}

	accountPrivateKey, err = getPropertyWithIndexFromFile(outputFile, "(2)", 1)
	if err != nil {
		return
	}

	if len(accountPrivateKey) < 3 {
		return errors.New("Len of account privite key is to small: " + accountPrivateKey)
	}

	accountPrivateKey = accountPrivateKey[2:len(accountPrivateKey)]

	identiyPrivateKey, err = getPropertyWithIndexFromFile(outputFile, "(0)", 1)
	if err != nil {
		return
	}

	return
}

func contractsAreDeployedUsingTruffle() (err error) {

	if environmentIsSet {
		return
	}

	command := ExecCommand{
		Command:   "./node_modules/.bin/truffle",
		Directory: platformContractsDir,
		Args:      []string{"compile"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	output := "migrate.out"
	command.Args = []string{"migrate", "--network", "local"}
	command.OutputFile = output
	err = runCommand(command)

	singnetTokenAddress, err = getPropertyFromFile(output, "SingularityNetToken:")
	if err != nil {
		return
	}

	registryAddress, err = getPropertyFromFile(output, "Registry:")
	if err != nil {
		return
	}

	agentFactoryAddress, err = getPropertyFromFile(output, "AgentFactory:")
	if err != nil {
		return
	}

	multiPartyEscrow, err = getPropertyFromFile(output, "MultiPartyEscrow:")
	if err != nil {
		return
	}

	return
}

func ipfsIsRunning(portAPI int, portGateway int) (err error) {

	if environmentIsSet {
		return
	}

	env := []string{"IPFS_PATH=" + envGoPath + "/ipfs"}

	command := ExecCommand{
		Command: "ipfs",
		Env:     env,
		Args:    []string{"init"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	command.Args = []string{"bootstrap", "rm", "--all"}
	err = runCommand(command)

	if err != nil {
		return
	}

	addressAPI := "/ip4/127.0.0.1/tcp/" + toString(portAPI)
	command.Args = []string{"config", "Addresses.API", addressAPI}
	err = runCommand(command)

	if err != nil {
		return
	}

	addressGateway := "/ip4/0.0.0.0/tcp/" + toString(portGateway)
	command.Args = []string{"config", "Addresses.Gateway", addressGateway}
	err = runCommand(command)

	if err != nil {
		return
	}

	outputFile = logPath + "/ipfs.log"
	command.OutputFile = outputFile
	command.Args = []string{"daemon"}
	err = runCommandAsync(command)

	if err != nil {
		return
	}

	outputContainsStrings = []string{
		"Daemon is ready",
		"server listening on " + addressAPI,
		"server listening on " + addressGateway,
	}
	exists, err := checkWithTimeout(5000, 500, checkFileContainsStrings)

	if err != nil {
		return
	}

	if !exists {
		return errors.New("Etherium networks is not started")
	}

	return nil
}

func identityIsCreatedWithUserAndPrivateKey(user string, privateKey string) (err error) {

	if environmentIsSet {
		return
	}

	command := ExecCommand{
		Command: "snet",
		Args:    []string{"identity", "create", user, "key", "--private-key", identiyPrivateKey},
	}
	err = runCommand(command)

	if err != nil {
		return
	}

	command.Args = []string{"identity", "snet-user"}
	return runCommand(command)
}

func snetIsConfiguredWithEthereumRPCEndpoint(endpointEthereumRPC int) (err error) {

	if environmentIsSet {
		return
	}

	config := `
[network.local]
default_eth_rpc_endpoint = http://localhost:` + toString(endpointEthereumRPC)

	err = appendToFile(snetConfigFile, config)

	if err != nil {
		return
	}

	command := ExecCommand{
		Command: "snet",
		Args:    []string{"network", "local"},
	}
	err = runCommand(command)

	if err != nil {
		return
	}

	outputFile = snetConfigFile
	outputContainsStrings = []string{"session"}
	exists, e := checkWithTimeout(5000, 500, checkFileContainsStrings)

	if !exists {
		return errors.New("snet config file is not created: " + snetConfigFile)
	}

	return e
}

func snetIsConfiguredWithIPFSEndpoint(endpointIPFS int) (err error) {

	if environmentIsSet {
		return
	}

	command := ExecCommand{
		Command: "sed",
		Args:    []string{"-ie", "/ipfs/,+2d", snetConfigFile},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	config := `
[ipfs]
default_ipfs_endpoint = http://localhost:` + toString(endpointIPFS)

	return appendToFile(snetConfigFile, config)
}

func organizationIsAdded(table *gherkin.DataTable) (err error) {

	if environmentIsSet {
		return
	}

	organization := getTableValue(table, "organization")

	args := []string{
		"contract", "Registry",
		"--at", registryAddress,
		"createOrganization", organization,
		"[\"" + organizationAddress + "\"]",
		"--transact",
		"--yes",
	}

	command := ExecCommand{
		Command: "snet",
		Args:    args,
	}

	err = runCommand(command)

	environmentIsSet = true

	return
}

func serviceIsRegistered(table *gherkin.DataTable, dir string) (err error) {

	if serviceIsPublished {
		return
	}

	name := getTableValue(table, "name")
	serviceSpec := getTableValue(table, "service_spec")
	price := getTableValue(table, "price")
	endpoint := getTableValue(table, "endpoint")
	tags := getTableValue(table, "tags")
	description := getTableValue(table, "description")

	command := ExecCommand{
		Command:   "snet",
		Directory: dir,
		Input:     []string{"", serviceSpec, name, "", price, endpoint, tags, description},
		Args:      []string{"service", "init"},
	}

	return runCommand(command)
}

func serviceIsPublishedToNetwork(dir string, serviceFile string) (err error) {

	if serviceIsPublished {
		return
	}

	args := []string{
		"service", "publish", "local",
		"--config", serviceFile,
		"--agent-factory-at", agentFactoryAddress,
		"--registry-at", registryAddress,
		"--yes",
	}

	command := ExecCommand{
		Command:   "snet",
		Directory: dir,
		Args:      args,
	}

	err = runCommand(command)

	if err != nil {
		return err
	}

	agentAddress, err = getPropertyFromFile(
		dir+"/"+serviceFile,
		"\"agentAddress\":",
	)

	if err != nil {
		return err
	}

	if len(agentAddress) < 2 {
		return errors.New("Len of accoagent address is to small: " + agentAddress)
	}

	agentAddress = agentAddress[1 : len(agentAddress)-1]

	serviceIsPublished = true

	return
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
