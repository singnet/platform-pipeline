package main

import (
	"errors"
	"log"

	"github.com/DATA-DOG/godog/gherkin"
)

var platformContractsDir string
var exampleServiceDir string
var dnnModelServicesDir string
var treasurerServerDir string
var snetConfigFile string

var treasurerPrivateKey string
var accountPrivateKey string
var identiyPrivateKey string

var snetIdentityAddress string
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
	treasurerServerDir = envSingnetRepos + "/treasurer"
	snetConfigFile = envHome + "/.snet/config"
}

func ethereumNetworkIsRunningOnPort(port int) (err error) {

	if environmentIsSet {
		return
	}

	outputFile := logPath + "/ganache.log"

	fileContains := checkFileContains{
		output:  outputFile,
		strings: []string{"Listening on 127.0.0.1:" + toString(port)},
	}

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

	exists, err := checkWithTimeout(5000, 500, checkFileContainsStringsFunc(fileContains))
	if err != nil {
		return
	}

	if !exists {
		return errors.New("Etherium networks is not started")
	}

	snetIdentityAddress, err = getPropertyFromFile(outputFile, "(0)")
	if err != nil {
		return
	}

	organizationAddress, err = getPropertyFromFile(outputFile, "(1)")
	if err != nil {
		return
	}

	treasurerPrivateKey, err = getPrivateKey("1", outputFile)
	if err != nil {
		return
	}

	accountPrivateKey, err = getPrivateKey("2", outputFile)
	if err != nil {
		return
	}

	identiyPrivateKey, err = getPropertyWithIndexFromFile(outputFile, "(0)", 1)
	if err != nil {
		return
	}

	return
}

func getPrivateKey(index string, file string) (key string, err error) {

	key, err = getPropertyWithIndexFromFile(file, "("+index+")", 1)
	if err != nil {
		return
	}

	if len(key) < 3 {
		err = errors.New("Len of account privite key is to small: " + key)
		return
	}

	key = key[2:len(key)]

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

	output := logPath + "/migrate.out"
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

	outputFile := logPath + "/ipfs.log"
	command.OutputFile = outputFile
	command.Args = []string{"daemon"}
	err = runCommandAsync(command)

	if err != nil {
		return
	}

	fileContains := checkFileContains{
		output: outputFile,
		strings: []string{
			"Daemon is ready",
			"server listening on " + addressAPI,
			"server listening on " + addressGateway,
		},
	}

	exists, err := checkWithTimeout(5000, 500, checkFileContainsStringsFunc(fileContains))

	if err != nil {
		return
	}

	if !exists {
		return errors.New("Etherium networks is not started")
	}

	return nil
}

func identityIsCreatedWithUser(user string) (err error) {

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

	fileContains := checkFileContains{
		output:  snetConfigFile,
		strings: []string{"session"},
	}

	exists, e := checkWithTimeout(5000, 500, checkFileContainsStringsFunc(fileContains))

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
