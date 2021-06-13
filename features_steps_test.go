package main

import (
	"github.com/cucumber/godog"
)

func FeatureContext(s *godog.ScenarioContext) {

	// background
	s.Step(`^Ethereum network is running on port (\d+)$`, ethereumNetworkIsRunningOnPort)
	s.Step(`^Contracts are deployed using Truffle$`, contractsAreDeployedUsingTruffle)
	s.Step(`^IPFS is running with API port (\d+) and Gateway port (\d+)$`, ipfsIsRunning)
	s.Step(`^snet is configured local rpc$`, snetIsConfiguredLocalRpc)
	s.Step(`^Organization is added$`, organizationIsAdded)

	// example-service-services sample
	s.Step(`^example-service service is registered$`, exampleserviceServiceIsRegistered)
	s.Step(`^example-service service snet-daemon config file is created$`, exampleserviceServiceSnetdaemonConfigFileIsCreated)
	s.Step(`^example-service service is running$`, exampleserviceServiceIsRunning)
	//todo mint function needs to be called

}
