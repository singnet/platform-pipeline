Feature: Publish example service 

Scenario: Run all services and publish example service 
	Given Ethereum network is running on port 8545 
	Given Contracts are deployed using Truffle 
	Given IPFS is running with API port 5002 and Gateway port 8081 
	Given Identity is created with user "snet-user" and private key "0xc71478a6d0fe44e763649de0a0deb5a080b788eefbbcf9c6f7aef0dd5dbd67e0" 
	Given snet is configured with Ethereum RPC endpoint 8545 
	Given snet is configured with IPFS endpoint 5002
	Given Organization is added:
		| organization | address | member |
		| ExampleOrganization | 0x4e74fefa82e83e0964f0d9f53c68e03f7298a8b2 | 0x3b2b3c2e2e7c93db335e69d827f3cc4bc2a2a2cb |
	Given  example-service is registered
		| name | price | endpoint | tags | description |
		| ExampleOrganization | 1 | http://localhost:8080 | example service | Example service | 
	Given  example-service is published to network
		| agent factory address | registry address |
		| 0x5c7a4290f6f8ff64c69eeffdfafc8644a4ec3a4e | 0x4e74fefa82e83e0964f0d9f53c68e03f7298a8b2 | 
