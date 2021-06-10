package main

import (
	"github.com/DATA-DOG/godog"
)

func FeatureContext(s *godog.Suite) {

	// background
	s.Step(`^Ethereum network is running on port (\d+)$`, ethereumNetworkIsRunningOnPort)
	s.Step(`^Contracts are deployed using Truffle$`, contractsAreDeployedUsingTruffle)
	s.Step(`^IPFS is running with API port (\d+) and Gateway port (\d+)$`, ipfsIsRunning)
	s.Step(`^snet is configured local rpc$`, snetIsConfiguredLocalRpc)
	s.Step(`^Organization is added:$`, organizationIsAdded)

	// example-service-services sample
	s.Step(`^example-service service is registered$`, exampleserviceServiceIsRegistered)
	s.Step(`^example-service service snet-daemon config file is created$`, exampleserviceServiceSnetdaemonConfigFileIsCreated)
	s.Step(`^example-service service is running$`, exampleserviceServiceIsRunning)
	//todo mint function needs to be called
	//s.Step(`^example-service make a call using payment channel$`, exampleserviceMakeACallUsingPaymentChannel)
	//s.Step(`^example-service claim channel by treasurer server$`, exampleserviceClaimChannelByTreasurerServer)
}
