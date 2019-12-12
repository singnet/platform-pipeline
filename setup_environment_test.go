package main

import (
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/ethereum/go-ethereum/common"
	"log"
)

var platformContractsDir string
var exampleServiceDir string
var treasurerServerDir string
var snetConfigFile string

var snetIdentityAddress string

var singnetTokenAddress string
var registryAddress string
var multiPartyEscrow string
var organizationAddress string

func init() {
	platformContractsDir = envSingnetRepos + "/platform-contracts"
	exampleServiceDir = envSingnetRepos + "/example-service"
	treasurerServerDir = envSingnetRepos + "/treasurer"
	snetConfigFile = envHome + "/.snet/config"
}

func ethereumNetworkIsRunningOnPort(port int) (err error) {

	output := logPath + "/ganache.log"
	NewCommand().Run("killall node || echo \"supress an error\"")
	NewCommand().Run("killall etcd || echo \"supress an error\"")
	err = NewCommand().Dir(platformContractsDir).
		Output(output).
		RunAsync("./node_modules/.bin/ganache-cli --deterministic --mnemonic \"gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce\"").
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
		Run("ipfs shutdown || echo \"supress an error\"").
		Run("rm -rf ~/.ipfs").
		Run("ipfs init").
		Run("ipfs bootstrap rm --all").
		Run("ipfs config Addresses.API %s", addressAPI).
		Run("ipfs config Addresses.Gateway %s", addressGateway).
		Output(outputFile).
		RunAsync("ipfs daemon").
		CheckOutput(
			"Daemon is ready",
			"server listening on "+addressAPI,
			"server listening on "+addressGateway).
		Err()

	return
}

func snetIsConfiguredLocalRpc(table *gherkin.DataTable) (err error) {
	rpc_port := getTableValue(table, "Ethereum RPC port")
	user_name := getTableValue(table, "user name")
	//ipfs_port := getTableValue(table, "IPFS port")

	err = NewCommand().
		//Run("rm -rf ~/.snet").
		Run("snet network create local http://localhost:%s", rpc_port).
		Run("snet identity create %s rpc --network local", user_name).
		//Run("snet set default_ipfs_endpoint http://localhost:%s", ipfs_port).
		Run("snet set current_singularitynettoken_at " + singnetTokenAddress).
		Run("snet set current_registry_at " + registryAddress).
		Run("snet set current_multipartyescrow_at " + multiPartyEscrow).
		Err()
	return
}

func organizationIsAdded(table *gherkin.DataTable) (err error) {

	organization := getTableValue(table, "organization")
	etcd_endpoint := getTableValue(table, "etcd endpoint")
	group_name := getTableValue(table, "group name")
	org_type := getTableValue(table, "type")

	//snet organization add-group group1 0x42A605c07EdE0E1f648aB054775D6D4E38496144 5.5.6.7:8089

	err = NewCommand().
		Run("snet organization metadata-init %s %s %s", organization, organization, org_type).
		Run("snet organization add-group %s `snet account print --wallet-index 1` %s ", group_name, etcd_endpoint).
		Run("snet organization create %s -y", organization).
		Err()

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
