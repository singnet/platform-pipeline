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

var singnetTokenAddress string
var registryAddress string
var multiPartyEscrow string
var organizationAddress string

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

	err = NewCommand().Dir(platformContractsDir).
		Output(outputFile).
		RunAsync("./node_modules/.bin/ganache-cli --mnemonic \"gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce\"").
		CheckOutput("Listening on 127.0.0.1:" + toString(port)).
		Err()

	if err != nil {
		return
	}

	return initContractAddresses(outputFile)
}

func initContractAddresses(outputFile string) (err error) {

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
		command:   "./node_modules/.bin/truffle",
		directory: platformContractsDir,
		args:      []string{"compile"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	output := logPath + "/migrate.out"
	command.args = []string{"migrate", "--network", "local"}
	command.outputFile = output
	err = runCommand(command)

	singnetTokenAddress, err = getPropertyFromFile(output, "SingularityNetToken:")
	if err != nil {
		return
	}

	registryAddress, err = getPropertyFromFile(output, "Registry:")
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
		command: "ipfs",
		env:     env,
		args:    []string{"init"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	command.args = []string{"bootstrap", "rm", "--all"}
	err = runCommand(command)

	if err != nil {
		return
	}

	addressAPI := "/ip4/127.0.0.1/tcp/" + toString(portAPI)
	command.args = []string{"config", "Addresses.API", addressAPI}
	err = runCommand(command)

	if err != nil {
		return
	}

	addressGateway := "/ip4/0.0.0.0/tcp/" + toString(portGateway)
	command.args = []string{"config", "Addresses.Gateway", addressGateway}
	err = runCommand(command)

	if err != nil {
		return
	}

	outputFile := logPath + "/ipfs.log"
	command.outputFile = outputFile
	command.args = []string{"daemon"}
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
		command: "snet",
		args:    []string{"identity", "create", user, "key", "--private-key", identiyPrivateKey},
	}
	err = runCommand(command)

	if err != nil {
		return
	}

	command.args = []string{"identity", "snet-user"}
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
		command: "snet",
		args:    []string{"network", "local"},
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
		command: "sed",
		args:    []string{"-ie", "/ipfs/,+2d", snetConfigFile},
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

	return NewCommand().
		Run("snet organization create %s --registry-at %s -y", organization, registryAddress).
		Err()
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
