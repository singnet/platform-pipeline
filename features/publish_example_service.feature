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

	Scenario: Publish example service
		When  example-service is registered
			| name                | service_spec | price | endpoint              | tags            | description     |
			| ExampleOrganization | service_spec | 1     | http://localhost:8080 | example service | Example service |
		When example-service is published to network
		When example-service is run with snet-daemon
			| daemon port | ethereum endpoint port | passthrough endpoint port |
			| 8080        | 8545                   | 5001                      |
		Then SingularityNET job is created
			| max price |
			| 100000000 |

	Scenario: Publish dnn-model-services
		When  dnn-model service is registered
			| name                | service_spec         | price | endpoint              | tags        | description         |
			| ExampleOrganization | service/service_spec | 1     | http://localhost:8090 | dnn service | DNN Example service |
		When  dnn-model service is published to network
		When  dnn-model mpe service is registered
			| name                | display name      | endpoint              | group  |
			| DNNModelService     | DNN Model Service | http://localhost:8090 | group1 |
		When  dnn-model service snet-daemon config file is created
            | daemon port | ethereum endpoint port | passthrough endpoint port | price |
            | 8090        | 8545                   | 7003                      | 10    |
		When dnn-model service is running
		When dnn-model open the payment channel
		When dnn-model compile protobuf
		Then dnn-model make a call using payment channel
		Then dnn-model claim channel by treasurer server
			| ethereum endpoint port |
			| 8545                   |
