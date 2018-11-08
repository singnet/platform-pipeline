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

	// example-service sample
	s.Step(`^example-service is registered$`, exampleserviceIsRegistered)
	s.Step(`^example-service is published to network$`, exampleserviceIsPublishedToNetwork)
	s.Step(`^example-service is run with snet-daemon$`, exampleserviceIsRunWithSnetdaemon)
	s.Step(`^SingularityNET job is created$`, singularityNETJobIsCreated)

	// dnn-model-services sample
	s.Step(`^dnn-model service is registered$`, dnnmodelServiceIsRegistered)
	s.Step(`^dnn-model service is published to network$`, dnnmodelServiceIsPublishedToNetwork)
	s.Step(`^dnn-model mpe service is registered$`, dnnmodelMpeServiceIsRegistered)
	s.Step(`^dnn-model service snet-daemon config file is created$`, dnnmodelServiceSnetdaemonConfigFileIsCreated)
	s.Step(`^dnn-model service is running$`, dnnmodelServiceIsRunning)
	s.Step(`^dnn-model open the payment channel$`, dnnmodelOpenThePaymentChannel)
	s.Step(`^dnn-model compile protobuf$`, dnnmodelCompileProtobuf)
	s.Step(`^dnn-model make a call using payment channel$`, dnnmodelMakeACallUsingPaymentChannel)
	s.Step(`^dnn-model claim channel by treasurer server$`, dnnmodelClaimChannelByTreasurerServer)
}
