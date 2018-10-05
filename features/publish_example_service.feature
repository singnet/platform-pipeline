Feature: Publish example service

	Scenario: Run all services and publish example service
		Given Ethereum network is running on port 8545
		Given Contracts are deployed using Truffle
		Given IPFS is running with API port 5002 and Gateway port 8081
		Given Identity is created with user "snet-user" and private key "0xc71478a6d0fe44e763649de0a0deb5a080b788eefbbcf9c6f7aef0dd5dbd67e0"
		Given snet is configured with Ethereum RPC endpoint 8545
		Given snet is configured with IPFS endpoint 5002
		When Organization is added:
			| organization        |
			| ExampleOrganization |
		When  example-service is registered
			| name                | price | endpoint              | tags            | description     |
			| ExampleOrganization | 1     | http://localhost:8080 | example service | Example service |
		When  example-service is published to network
		When example-service is run with snet-daemon
			| daemon port | ethereum endpoint port | passthrough endpoint port |
			| 8080        | 8545                   | 5001                      |
		Then SingularityNET job is created
			| max price |
			| 100000000 |
