Feature: Publish example service

    Background: Run all services
		Given Ethereum network is running on port 8545
		Given Contracts are deployed using Truffle
		Given IPFS is running with API port 5002 and Gateway port 8081
        Given snet is configured local rpc
            | Ethereum RPC port | user name   | IPFS port |
            | 8545              | "snet-user" | 5002      |
		Given Organization is added:
			| organization        |
			| ExampleOrganization |

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
