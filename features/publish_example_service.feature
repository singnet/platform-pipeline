Feature: Publish example service

	Scenario: Run all services and publish example service
		Given Ethereum network is running on port 8545
		Given Contracts are deployed using Truffle
		Given IPFS is running with API port 5002 and Gateway port 8081
		Given Identity is created with user "snet-user" and private key "0xc71478a6d0fe44e763649de0a0deb5a080b788eefbbcf9c6f7aef0dd5dbd67e0"
		Given snet is configured with Ethereum RPC endpoint 8545
		Given snet is configured with IPFS endpoint 5002
		When Organization is added:
			| organization        | address                                    | member                                     |
			| ExampleOrganization | 0x8d1c8634f032d1c65c540faca15f7df83fbb9f8c | 0x3b2b3c2e2e7c93db335e69d827f3cc4bc2a2a2cb |
		When  example-service is registered
			| name                | price | endpoint              | tags            | description     |
			| ExampleOrganization | 1     | http://localhost:8080 | example service | Example service |
		When  example-service is published to network
			| agent factory address                      | registry address                           |
			| 0x8d1c8634f032d1c65c540faca15f7df83fbb9f8c | 0x8d1c8634f032d1c65c540faca15f7df83fbb9f8c |
		When example-service is run with snet-daemon
			| daemon port | ethereum endpoint port | passthrough endpoint port | agent contract address                     | private key                                                      |
			| 8080        | 8545                   | 5001                      | 0x3b07411493C72c5aEC01b6Cf3cd0981cF0586fA7 | ba398df3130586b0d5e6ef3f757bf7fe8a1299d4b7268fdaae415952ed30ba87 |
		Then SingularityNET job is created
			| max price | agent contract address                     |
			| 100000000 | 0x3b07411493C72c5aEC01b6Cf3cd0981cF0586fA7 |
