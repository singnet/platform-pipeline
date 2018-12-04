package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DATA-DOG/godog/gherkin"
)

const (
	configServiceName = "snetd.config.json"
)

func dnnmodelServiceIsRegistered(table *gherkin.DataTable) (err error) {

	name := getTableValue(table, "name")
	displayName := getTableValue(table, "display name")
	daemonPort := getTableValue(table, "daemon port")
	organization := getTableValue(table, "organization name")

	metadata := dnnModelServicesDir + "/service_metadata.json"
	cmd := NewCommand().Dir(dnnModelServicesDir)

	cmd.
		Run("snet service metadata_init service/service_spec \"%s\" %s",
			displayName, organizationAddress).
		CheckFileContains(metadata, "display_name", displayName).
		Run("snet service metadata_set_fixed_price 0.1").
		CheckFileContains(metadata, "fixed_price", "price_in_cogs", "10000000").
		Run("snet service metadata_add_endpoints localhost:%s", daemonPort).
		Run("snet service publish %s %s -y", organization, name)

	return cmd.Err()
}

func dnnmodelServiceSnetdaemonConfigFileIsCreated(table *gherkin.DataTable) (err error) {

	serviceName := getTableValue(table, "name")
	organizationName := getTableValue(table, "organization name")
	daemonPort := getTableValue(table, "daemon port")
	price := getTableValue(table, "price")

	snetdConfigTemplate := `
	{
		"SERVICE_NAME": "%s",
		"ORGANIZATION_NAME": "%s",
		"DAEMON_LISTENING_PORT": %s,
		"DAEMON_END_POINT": "localhost:%s",
		"ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:8545",
		"PASSTHROUGH_ENABLED": true,
		"PASSTHROUGH_ENDPOINT": "http://localhost:7003",
		"IPFS_END_POINT": "http://localhost:5002",
		"REGISTRY_ADDRESS_KEY": "%s",
		"price_per_call": %s,
		"log": {
		  "level": "debug",
		  "output": {
			"type": "stdout"
		  }
		},
		"payment_channel_storage_type": "etcd",
		"payment_channel_storage_client": {
		  "endpoints": [
			"http://127.0.0.1:2479"
		  ]
		},
		"payment_channel_storage_server": {
		  "host": "127.0.0.1",
		  "client_port": 2479,
		  "peer_port": 2480,
		  "token": "unique-token-dnn",
		  "cluster": "storage-1=http://127.0.0.1:2480",
		  "enabled": true
		}
	  }`
	snetdConfig := fmt.Sprintf(
		snetdConfigTemplate,
		serviceName,
		organizationName,
		daemonPort,
		daemonPort,
		registryAddress,
		price,
	)

	file := dnnModelServicesDir + "/" + configServiceName
	log.Printf("create snetd config: %s\n---\n:%s\n---\n", file, snetdConfig)

	return writeToFile(file, snetdConfig)
}

func dnnmodelServiceIsRunning() (err error) {

	err = os.Chmod(dnnModelServicesDir+"/buildproto.sh", 0544)

	if err != nil {
		return
	}

	output := logPath + "/dnn-model-service.log"
	cmd := NewCommand().Dir(dnnModelServicesDir)
	cmd.
		Run("./buildproto.sh").
		Output(output).
		RunAsync("python3 run_basic_service.py").
		CheckOutput("starting daemon")

	return cmd.Err()
}

func dnnmodelMakeACallUsingPaymentChannel(table *gherkin.DataTable) (err error) {

	name := getTableValue(table, "name")
	organization := getTableValue(table, "organization name")
	daemonPort := getTableValue(table, "daemon port")

	cmd := NewCommand().Dir(dnnModelServicesDir)
	cmd.
		Run("snet client balance").
		Run("snet client deposit 42000.22 -y").
		Run("snet client open_init_channel_registry %s %s 42 100000000 -y", organization, name).
		Run("snet client call 0 0.1 localhost:%s add '{\"a\":10,\"b\":32}'", daemonPort)

	return cmd.Err()
}

func dnnmodelClaimChannelByTreasurerServer(table *gherkin.DataTable) (err error) {

	err = os.Mkdir(treasurerServerDir, 0700)

	if err != nil {
		return
	}

	serviceName := getTableValue(table, "name")
	organizationName := getTableValue(table, "organization name")
	daemonPort := getTableValue(table, "daemon port")

	snetdConfigTemplate := `
	{
		"SERVICE_NAME": "%s",
		"ORGANIZATION_NAME": "%s",
		"DAEMON_LISTENING_PORT": %s,
		"DAEMON_END_POINT": "localhost:%s",
		"ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:8545",
		"PASSTHROUGH_ENABLED": true,
		"PASSTHROUGH_ENDPOINT": "http://localhost:7003",
		"IPFS_END_POINT": "http://localhost:5002",
		"REGISTRY_ADDRESS_KEY": "%s",
		"PRIVATE_KEY": "%s",
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
		serviceName,
		organizationName,
		daemonPort,
		daemonPort,
		registryAddress,
		treasurerPrivateKey,
	)

	log.Println("conf file:\n", snetdConfig)

	file := treasurerServerDir + "/snetd.config.json"
	err = writeToFile(file, snetdConfig)

	if err != nil {
		return
	}

	cmd := NewCommand().Dir(treasurerServerDir)
	cmd.
		Run("snetd list channels").
		Run("snetd claim --channel-id 0").
		Run("snet client balance"+
			" --account %s"+
			" --snt %s"+
			" --multipartyescrow %s",
			organizationAddress, singnetTokenAddress, multiPartyEscrow)

	return cmd.Err()
}
