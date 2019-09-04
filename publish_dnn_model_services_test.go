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

func exampleserviceServiceIsRegistered(table *gherkin.DataTable) (err error) {

	name := getTableValue(table, "name")
	displayName := getTableValue(table, "display name")
	daemonPort := getTableValue(table, "daemon port")
	organization := getTableValue(table, "organization name")
	groupname := getTableValue(table, "group name")

	metadata := exampleServiceDir + "/service_metadata.json"
	cmd := NewCommand().Dir(exampleServiceDir)
	cmd.
		Run("snet service metadata-init service/service_spec \"%s\" --encoding proto --service-type grpc --group-name %s ",
			displayName, groupname).
		CheckFileContains(metadata, "display_name", displayName).
		CheckFileContains(metadata, "group_name", "default_group").
		Run("snet service metadata-set-fixed-price %s 0.1", groupname).
		CheckFileContains(metadata, "fixed_price", "price_in_cogs", "10000000").
		Run("snet service metadata-add-endpoints default_group localhost:%s", daemonPort).
		Run("snet service publish %s %s -y", organization, name)

	return cmd.Err()
}

func exampleserviceServiceSnetdaemonConfigFileIsCreated(table *gherkin.DataTable) (err error) {

	serviceName := getTableValue(table, "name")
	organizationName := getTableValue(table, "organization name")
	daemonPort := getTableValue(table, "daemon port")

	snetdConfigTemplate := `
	{
        "payment_channel_storage_server": {
		"enabled": true
        },
		"SERVICE_ID": "%s",
		"ORGANIZATION_ID": "%s",
        "DAEMON_END_POINT": "localhost:%s",
		"ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:8545",
		"PASSTHROUGH_ENABLED": true,
		"PASSTHROUGH_ENDPOINT": "http://localhost:7003",
		"IPFS_END_POINT": "http://localhost:5002",
		"REGISTRY_ADDRESS_KEY": "%s",
        "pvt_key_for_metering": "efed2ea91d5ace7f9f7bd91e21223cfded31a6e3f1a746bc52821659e0c94e17",
        "metering_end_point":"http://demo8325345.mockable.io",
		"log": {
		  "level": "debug",
		  "output": {
			"type": "stdout"
		  }
		}
        
	  }`
	snetdConfig := fmt.Sprintf(
		snetdConfigTemplate,
		serviceName,
		organizationName,
		daemonPort,
		registryAddress,
	)

	file := exampleServiceDir + "/" + configServiceName
	log.Printf("create snetd config: %s\n---\n:%s\n---\n", file, snetdConfig)

	return writeToFile(file, snetdConfig)
}

func exampleserviceServiceIsRunning() (err error) {

	err = os.Chmod(exampleServiceDir+"/buildproto.sh", 0544)

	if err != nil {
		return
	}

	output := logPath + "/example-service.log"
	exampleRunCmd := "python3 run_example_service.py --daemon-config " + exampleServiceDir + "/" + configServiceName
	cmd := NewCommand().Dir(exampleServiceDir)
	cmd.Run("rm -rf storage-data-dir-1.etcd").
		Run("./buildproto.sh").
		Output(output).
		RunAsync(exampleRunCmd)
		//CheckOutput("starting daemon") //todo this does not work well and was failing the test cases

	return cmd.Err()
}

func exampleserviceMakeACallUsingPaymentChannel(table *gherkin.DataTable) (err error) {

	group_name := getTableValue(table, "group name")
	organization := getTableValue(table, "organization name")
	service := getTableValue(table, "service name")

	cmd := NewCommand().Dir(exampleServiceDir)
	cmd.
		Run("snet account balance").
		Run("snet account deposit 42000.22 -y").
		Run("snet channel open-init %s %s 42 +30days -y", organization, group_name).
		Run("snet --print-traceback client call %s %s %s add '{\"a\":10,\"b\":32}' -y", organization, service, group_name)

	return cmd.Err()
}

func exampleserviceClaimChannelByTreasurerServer(table *gherkin.DataTable) (err error) {

	daemonPort := getTableValue(table, "daemon port")
	cmd := NewCommand().Dir(exampleServiceDir)
	cmd.Run("snet treasurer print-unclaimed --endpoint localhost:%s --wallet-index 1", daemonPort).
		Run("snet treasurer claim-all --endpoint localhost:%s  --wallet-index 1 -y", daemonPort)
	return cmd.Err()
}
