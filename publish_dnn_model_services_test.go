package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

const (
	serviceName = "basic_service_one"
)

func dnnmodelServiceIsRegistered(table *gherkin.DataTable) (err error) {
	err = serviceIsRegistered(table, dnnModelServicesDir)
	return
}

func dnnmodelServiceIsPublishedToNetwork() (err error) {
	err = serviceIsPublishedToNetwork(dnnModelServicesDir, "./service.json")
	return
}

func dnnmodelMpeServiceIsRegistered(table *gherkin.DataTable) (err error) {

	name := getTableValue(table, "name")
	displayName := getTableValue(table, "display name")
	group := getTableValue(table, "group")
	endpoint := getTableValue(table, "endpoint")

	log.Println("dnnModelServicesDir: ", dnnModelServicesDir)

	output := dnnModelServicesDir + "/output.txt"

	// snet mpe-service publish_proto
	command := ExecCommand{
		Command:    "snet",
		Directory:  dnnModelServicesDir,
		Args:       []string{"mpe-service", "publish_proto", "service/service_spec/"},
		OutputFile: output,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	modelIpfsHash, err := readFile(output)
	log.Println("modelIpfsHash: ", modelIpfsHash)

	if err != nil {
		return
	}

	//snet mpe-service metadata_init
	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args: []string{
			"mpe-service", "metadata_init",
			modelIpfsHash, multiPartyEscrow, displayName,
		},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	// snet mpe-service metadata_add_group
	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args:      []string{"mpe-service", "metadata_add_group", group, organizationAddress},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	// snet mpe-service metadata_add_endpoints
	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args:      []string{"mpe-service", "metadata_add_endpoints", group, endpoint},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	// snet mpe-service  publish_service
	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args:      []string{"mpe-service", "publish_service", registryAddress, name, "Basic_Template", "-y"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	return
}

func dnnmodelServiceSnetdaemonConfigFileIsCreated(table *gherkin.DataTable) (err error) {

	daemonPort := getTableValue(table, "daemon port")
	price := getTableValue(table, "price")
	ethereumEndpointPort := getTableValue(table, "ethereum endpoint port")
	passthroughEndpointPort := getTableValue(table, "passthrough endpoint port")

	snetdConfigTemplate := `
	{
		"AGENT_CONTRACT_ADDRESS": "%s",
		"MULTI_PARTY_ESCROW_CONTRACT_ADDRESS": "%s",
		"PRIVATE_KEY": "%s",
		"DAEMON_LISTENING_PORT": %s,
		"ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:%s",
		"PASSTHROUGH_ENABLED": true,
		"PASSTHROUGH_ENDPOINT": "http://localhost:%s",
		"price_per_call": %s,
		"log": {
			"level": "debug",
			"output": {
				"type": "stdout"
			}
		},
		"payment_channel_storage_type": "etcd",
		"payment_channel_storage_client": {
			"endpoints": ["http://127.0.0.1:2479"]
		},
		"payment_channel_storage_server": {
			"host" : "127.0.0.1",
			"client_port": 2479,
			"peer_port": 2480,
			"token": "unique-token-dnn",
			"cluster": "storage-1=http://127.0.0.1:2480",
			"enabled": "true"
		}
	}
	`

	snetdConfig := fmt.Sprintf(
		snetdConfigTemplate,
		agentAddress,
		multiPartyEscrow,
		accountPrivateKey,
		daemonPort,
		ethereumEndpointPort,
		passthroughEndpointPort,
		price,
	)

	file := fmt.Sprintf("%s/snetd_%s_config.json", dnnModelServicesDir, serviceName)
	log.Printf("create snetd config: %s\n---\n:%s\n---\n", file, snetdConfig)

	err = writeToFile(file, snetdConfig)

	return
}

func dnnmodelServiceIsRunning() (err error) {

	err = os.Chmod(dnnModelServicesDir+"/buildproto.sh", 0544)

	if err != nil {
		return
	}

	command := ExecCommand{
		Command:   dnnModelServicesDir + "/buildproto.sh",
		Directory: dnnModelServicesDir,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	fileContains := checkFileContains{
		output:  logPath + "/dnn-model-services-" + serviceName + ".log",
		strings: []string{"multi_party_escrow_contract_address"},
	}

	command = ExecCommand{
		Command:    "python3",
		Directory:  dnnModelServicesDir,
		Args:       []string{"run_basic_service.py", "--daemon-config-path", "."},
		OutputFile: fileContains.output,
	}

	err = runCommandAsync(command)

	if err != nil {
		return err
	}

	_, err = checkWithTimeout(5000, 500, checkFileContainsStringsFunc(fileContains))

	return
}

func dnnmodelOpenThePaymentChannel() (err error) {

	command := ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args: []string{
			"contract",
			"SingularityNetToken", "--at", singnetTokenAddress,
			"approve", multiPartyEscrow, "1000000",
			"--transact",
			"-y",
		},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args: []string{
			"contract",
			"MultiPartyEscrow", "--at", multiPartyEscrow,
			"deposit", "1000000",
			"--transact",
			"-y",
		},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	output := dnnModelServicesDir + "/expiration.txt"

	command = ExecCommand{
		Command:    "snet",
		Directory:  dnnModelServicesDir,
		Args:       []string{"mpe-client", "block_number"},
		OutputFile: output,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	expirationText, err := readFile(output)
	if err != nil {
		return
	}

	expiration, err := strconv.Atoi(strings.TrimSpace(expirationText))

	if err != nil {
		return
	}

	expiration += 12000

	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args: []string{
			"contract",
			"MultiPartyEscrow", "--at", multiPartyEscrow,
			"openChannel", organizationAddress,
			"420000", strconv.Itoa(expiration), "0",
			"--transact",
			"-y",
		},
	}

	err = runCommand(command)

	return
}

func dnnmodelCompileProtobuf() (err error) {

	command := ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args: []string{
			"mpe-client",
			"compile_from_file",
			envSingnetRepos + "/dnn-model-services/Services/gRPC/Basic_Template/service/service_spec",
			"basic_tamplate_rpc.proto",
			"0",
		},
	}

	err = runCommand(command)

	return
}

func dnnmodelMakeACallUsingPaymentChannel() (err error) {

	outputFile := dnnModelServicesDir + "/output.txt"

	fileContains := checkFileContains{
		output:     outputFile,
		strings:    []string{organizationAddress, "420000"},
		ignoreCase: true,
	}

	command := ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args: []string{
			"mpe-client",
			"print_my_channels", multiPartyEscrow,
		},
		OutputFile: outputFile,
	}

	err = runCommand(command)

	ok, err := checkFileContainsStrings(fileContains)
	err = fileContainsError(fileContains, ok, err)

	if err != nil {
		return
	}

	for i := 0; i < 3; i++ {
		command = ExecCommand{
			Command:   "snet",
			Directory: dnnModelServicesDir,
			Args: []string{
				"mpe-client",
				"call_server", multiPartyEscrow,
				"0", "10", "localhost:8090", "Addition", "add", `{"a":10,"b":32}`,
			},
		}

		err = runCommand(command)
		if err != nil {
			return
		}

	}

	return
}

func dnnmodelClaimChannelByTreasurerServer(table *gherkin.DataTable) (err error) {

	err = os.Mkdir(treasurerServerDir, 0700)

	if err != nil {
		return
	}

	ethereumEndpointPort := getTableValue(table, "ethereum endpoint port")

	snetdConfigTemplate := `
	{
		"AGENT_CONTRACT_ADDRESS": "%s",
		"MULTI_PARTY_ESCROW_CONTRACT_ADDRESS": "%s",
		"PRIVATE_KEY": "%s",
		"ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:%s",
		"log": {
			"level": "debug",
			"output": {
				"type": "stdout"
			}
		},
		"payment_channel_storage_type": "etcd",
		"payment_channel_storage_client": {
			"connection_timeout": "5s",
			"request_timeout": "3s",
			"endpoints": ["http://127.0.0.1:2479"]
		}
	}`

	snetdConfig := fmt.Sprintf(
		snetdConfigTemplate,
		agentAddress,
		multiPartyEscrow,
		treasurerPrivateKey,
		ethereumEndpointPort,
	)

	log.Println("conf file:\n", snetdConfig)

	file := treasurerServerDir + "/snetd.config.json"
	err = writeToFile(file, snetdConfig)

	if err != nil {
		return
	}

	output := treasurerServerDir + "/output.txt"

	// snet contract MultiPartyEscrow
	command := ExecCommand{
		Command:   "snet",
		Directory: treasurerServerDir,
		Args: []string{
			"contract",
			"MultiPartyEscrow", "--at", multiPartyEscrow,
			"channels", "0",
		},
		OutputFile: output,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	fileContains := checkFileContains{
		output:     output,
		strings:    []string{snetIdentityAddress, organizationAddress, "420000"},
		ignoreCase: true,
	}

	ok, err := checkFileContainsStrings(fileContains)
	err = fileContainsError(fileContains, ok, err)

	if err != nil {
		return
	}

	// snetd claim
	command = ExecCommand{
		Command:   "snetd",
		Directory: treasurerServerDir,
		Args:      []string{"claim", "--channel-id", "0"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	// snet contract MultiPartyEscrow
	command = ExecCommand{
		Command:   "snet",
		Directory: treasurerServerDir,
		Args: []string{
			"contract",
			"MultiPartyEscrow", "--at", multiPartyEscrow,
			"channels", "0",
		},
		OutputFile: output,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	fileContains = checkFileContains{
		output:     output,
		strings:    []string{snetIdentityAddress, organizationAddress, "419970"},
		ignoreCase: true,
	}

	ok, err = checkFileContainsStrings(fileContains)
	err = fileContainsError(fileContains, ok, err)

	if err != nil {
		return
	}

	return
}
