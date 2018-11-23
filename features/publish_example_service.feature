Feature: Publish example service

    Background: Run all services
		Given Ethereum network is running on port 8545
		Given Contracts are deployed using Truffle
		Given IPFS is running with API port 5002 and Gateway port 8081
		Given Identity is created with user "snet-user"
		Given snet is configured with Ethereum RPC endpoint 8545
		Given snet is configured with IPFS endpoint 5002
		Given Organization is added:
			| organization        |
			| ExampleOrganization |

	# Scenario: Publish example service
	# 	When  example-service is registered
	# 		| name                | service_spec | price | endpoint              | tags            | description     |
	# 		| ExampleOrganization | service_spec | 1     | http://localhost:8080 | example service | Example service |
	# 	When example-service is published to network
	# 	When example-service is run with snet-daemon
	# 		| daemon port | ethereum endpoint port | passthrough endpoint port |
	# 		| 8080        | 8545                   | 5001                      |
	# 	Then SingularityNET job is created
	# 		| max price |
	# 		| 100000000 |

	Scenario: Publish dnn-model-services
		When  dnn-model service is registered
			| name            | display name      | organization name   | daemon port |
			| DNNModelService | DNN Model Service | ExampleOrganization | 8090        |
		When  dnn-model service snet-daemon config file is created
            | name            | organization name   | daemon port | price |
            | DNNModelService | ExampleOrganization | 8090        | 10    |
		When dnn-model service is running
		Then dnn-model make a call using payment channel
            | name            | organization name   | daemon port |
            | DNNModelService | ExampleOrganization | 8090        |
		Then dnn-model claim channel by treasurer server
            | name            | organization name   | daemon port |
            | DNNModelService | ExampleOrganization | 8090        |
