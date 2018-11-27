package main

import (
	"github.com/DATA-DOG/godog"
)

func FeatureContext(s *godog.Suite) {

	// background
	s.Step(`^Ethereum network is running on port (\d+)$`, ethereumNetworkIsRunningOnPort)
	s.Step(`^Contracts are deployed using Truffle$`, contractsAreDeployedUsingTruffle)
	s.Step(`^IPFS is running with API port (\d+) and Gateway port (\d+)$`, ipfsIsRunning)
	s.Step(`^Identity is created with user "([^"]*)"$`, identityIsCreatedWithUser)
	s.Step(`^snet is configured with Ethereum RPC endpoint (\d+)$`, snetIsConfiguredWithEthereumRPCEndpoint)
	s.Step(`^snet is configured with IPFS endpoint (\d+)$`, snetIsConfiguredWithIPFSEndpoint)
	s.Step(`^Organization is added:$`, organizationIsAdded)

	// dnn-model-services sample
	s.Step(`^dnn-model service is registered$`, dnnmodelServiceIsRegistered)
	s.Step(`^dnn-model service snet-daemon config file is created$`, dnnmodelServiceSnetdaemonConfigFileIsCreated)
	s.Step(`^dnn-model service is running$`, dnnmodelServiceIsRunning)
	s.Step(`^dnn-model make a call using payment channel$`, dnnmodelMakeACallUsingPaymentChannel)
	s.Step(`^dnn-model claim channel by treasurer server$`, dnnmodelClaimChannelByTreasurerServer)
}
