package main

import (
	"errors"
	"log"

	"github.com/ethereum/go-ethereum/common"
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

func init() {
	platformContractsDir = envSingnetRepos + "/platform-contracts"
	exampleServiceDir = envSingnetRepos + "/example-service"
	dnnModelServicesDir = envSingnetRepos + "/dnn-model-services/Services/gRPC/Basic_Template"
	treasurerServerDir = envSingnetRepos + "/treasurer"
	snetConfigFile = envHome + "/.snet/config"
}

func ethereumNetworkIsRunningOnPort(port int) (err error) {

	output := logPath + "/ganache.log"

	err = NewCommand().Dir(platformContractsDir).
		Output(output).
		RunAsync("./node_modules/.bin/ganache-cli --mnemonic \"gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce\"").
		CheckOutput("Listening on 127.0.0.1:" + toString(port)).
		Err()

	if err != nil {
		return
	}

	return initAddresses(output)
}

func toChecksumAddress(hexAddress string) string {
	address := common.HexToAddress(hexAddress)
	mixedAddress := common.NewMixedcaseAddress(address)
	return mixedAddress.Address().String()
}

func initAddresses(output string) (err error) {

	snetIdentityAddress, err = getPropertyFromFile(output, "(0)")
	if err != nil {
		return
	}

	snetIdentityAddress = toChecksumAddress(snetIdentityAddress)

	organizationAddress, err = getPropertyFromFile(output, "(1)")
	if err != nil {
		return
	}
	organizationAddress = toChecksumAddress(organizationAddress)

	treasurerPrivateKey, err = getPrivateKey("1", output)
	if err != nil {
		return
	}

	accountPrivateKey, err = getPrivateKey("2", output)
	if err != nil {
		return
	}

	identiyPrivateKey, err = getPropertyWithIndexFromFile(output, "(0)", 1)
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

	key = key[2:]

	return
}

func contractsAreDeployedUsingTruffle() (err error) {

	output := logPath + "/migrate.out"

	err = NewCommand().
		Dir(platformContractsDir).
		Run("./node_modules/.bin/truffle compile").
		Output(output).
		Run("./node_modules/.bin/truffle migrate --network local").
		Err()

	if err != nil {
		return
	}

	return initContractAddresses(output)
}

func initContractAddresses(output string) (err error) {

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

	addressAPI := "/ip4/127.0.0.1/tcp/" + toString(portAPI)
	addressGateway := "/ip4/0.0.0.0/tcp/" + toString(portGateway)
	outputFile := logPath + "/ipfs.log"

	err = NewCommand().
		Run("ipfs init").
		Run("ipfs bootstrap rm --all").
		Run("ipfs config Addresses.API %s", addressAPI).
		Run("ipfs config Addresses.Gateway %s", addressGateway).
		Output(outputFile).
		RunAsync("ipfs daemon").
		CheckOutput(
		"Daemon is ready",
		"server listening on " + addressAPI,
		"server listening on " + addressGateway).
		Err()


	return
}

func identityIsCreatedWithUser(user string) (err error) {

	err = NewCommand().
		Run("snet identity create %s key --private-key %s", user, identiyPrivateKey).
		Run("snet identity snet-user").
		Err()

	return

}

func snetIsConfiguredWithEthereumRPCEndpoint(endpointEthereumRPC int) (err error) {

	config := `
[network.local]
default_eth_rpc_endpoint = http://localhost:` + toString(endpointEthereumRPC)

	err = appendToFile(snetConfigFile, config)

	if err != nil {
		return
	}

	err = NewCommand().
		Run("snet network local").
		CheckFileContains(snetConfigFile, "session").
		Err()

	return
}

func snetIsConfiguredWithIPFSEndpoint(endpointIPFS int) (err error) {

	err = NewCommand().
		Run("sed -ie '/ipfs/,+2d' %s", snetConfigFile).
		Err()

	if err != nil {
		return
	}

	config := `
[ipfs]
default_ipfs_endpoint = http://localhost:` + toString(endpointIPFS)

	return appendToFile(snetConfigFile, config)
}

func organizationIsAdded(table *gherkin.DataTable) (err error) {

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